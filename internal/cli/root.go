package cli

import (
	"context"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/hrodrig/kiko/internal/config"
	"github.com/hrodrig/kiko/internal/hit"
	"github.com/hrodrig/kiko/internal/log"
	"github.com/hrodrig/kiko/internal/server"
	"github.com/hrodrig/kiko/internal/store"
	"github.com/hrodrig/kiko/internal/version"
	"github.com/hrodrig/kiko/internal/visitor"
	"github.com/spf13/cobra"
)

func Execute() int {
	var cfgFile string

	root := &cobra.Command{
		Use:     "kiko",
		Short:   "Privacy-first web analytics collector",
		Version: version.Version,
	}

	root.PersistentFlags().StringVarP(&cfgFile, "config", "c", "", "config file (default ./kiko.yml)")

	serve := &cobra.Command{
		Use:   "serve",
		Short: "Start the analytics server",
		RunE: func(cmd *cobra.Command, args []string) error {
			return serveCmd(cfgFile)
		},
	}
	serve.Flags().String("listen", ":8080", "HTTP listen address")
	root.AddCommand(serve)

	versionCmd := &cobra.Command{
		Use:   "version",
		Short: "Print version info",
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Println(version.BuildInfo())
		},
	}
	root.AddCommand(versionCmd)

	if err := root.Execute(); err != nil {
		return 1
	}
	return 0
}

func serveCmd(cfgFile string) error {
	cfg, err := config.Load(cfgFile)
	if err != nil {
		return err
	}

	_, _, handler, cleanup, err := initRuntime(cfg)
	if err != nil {
		return err
	}
	defer cleanup()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	srv := &http.Server{
		Addr:    cfg.Listen,
		Handler: handler,
	}

	cfg.Log.Info("kiko v%s starting on %s (db=%s, flush=%ds, cap=%d, log=%s)",
		version.Version, cfg.Listen, cfg.Database.NormalizedDriver(),
		cfg.Buffer.FlushInterval, cfg.Buffer.Capacity, cfg.Log.LevelName())
	if cfg.Database.NormalizedDriver() == "sqlite" {
		cfg.Log.Info("database path: %s", cfg.Database.Path)
	}
	cfg.Log.Info("public_url: %s", cfg.PublicURL)

	errCh := make(chan error, 1)
	go func() {
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- err
		}
		close(errCh)
	}()

	select {
	case err := <-errCh:
		return err
	case <-ctx.Done():
		cfg.Log.Info("shutting down")
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	return srv.Shutdown(shutdownCtx)
}

func initRuntime(cfg *config.Config) (store.Store, hit.Buffer, http.Handler, func(), error) {
	st, err := store.Open(cfg.Database)
	if err != nil {
		return nil, nil, nil, nil, err
	}

	buf := hit.NewBuffer(cfg.Buffer.Capacity)
	flushCtx, cancel := context.WithCancel(context.Background())
	go runFlusher(flushCtx, st, buf, cfg)

	vh := visitor.NewHasher(cfg.Visitor.Salt)
	if vh.DevSalt() {
		cfg.Log.Warn("visitor.salt not set — using dev default; set KIKO_VISITOR_SALT in production")
	}

	var rl *server.RateLimiter
	if cfg.RateLimit.Enabled {
		rl = server.NewRateLimiter(server.RateLimitConfig{
			RequestsPerSec: cfg.RateLimit.RequestsPerSec,
			Burst:          cfg.RateLimit.Burst,
			TrustProxy:     cfg.Filter.TrustProxy,
		})
	}

	filter, err := server.NewHitFilter(server.FilterConfig{
		AllowedHosts:       cfg.AllowedHosts,
		IgnoreIPs:          cfg.Filter.IgnoreIPs,
		TrustProxy:         cfg.Filter.TrustProxy,
		BlockDatacenterIPs: cfg.Filter.BlockDatacenterIPs,
		DatacenterExtra:    cfg.Filter.DatacenterCIDRs,
	})
	if err != nil {
		cancel()
		st.Close()
		return nil, nil, nil, nil, err
	}

	hostRL := server.NewHostRateLimiter(cfg.RateLimit.HostRequestsPerSec, cfg.RateLimit.HostBurst)

	var apiRL *server.APIRateLimiter
	if cfg.API.RateLimit.Enabled {
		apiRL = server.NewAPIRateLimiter(cfg.API.RateLimit.RequestsPerSec, cfg.API.RateLimit.Burst)
	}
	if cfg.API.Key == "" {
		cfg.Log.Warn("api.key not set — stats endpoints are open; set KIKO_API_KEY in production")
	}

	opts := []server.ServerOption{
		server.WithStats(server.StatsConfig{APIKey: cfg.API.Key}, apiRL),
		server.WithIngest(filter, hostRL, cfg.Filter.TrustProxy),
	}
	sv := server.New(st, buf, cfg.Log, vh, rl, opts...)
	cleanup := func() {
		cancel()
		flushHits(st, buf, cfg.Log)
		if rl != nil {
			rl.Shutdown()
		}
		st.Close()
	}
	return st, buf, sv.Handler(), cleanup, nil
}

func runFlusher(ctx context.Context, st store.Store, buf hit.Buffer, cfg *config.Config) {
	ticker := time.NewTicker(time.Duration(cfg.Buffer.FlushInterval) * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			flushHits(st, buf, cfg.Log)
		}
	}
}

func flushHits(st store.Store, buf hit.Buffer, l *log.Logger) {
	hits := buf.Flush()
	if len(hits) == 0 {
		return
	}
	if err := st.SaveHits(hits); err != nil {
		l.Error("flush failed (%d hits): %v", len(hits), err)
		return
	}
	l.Debug("flushed %d hits", len(hits))
	if drops := buf.Drops(); drops > 0 {
		l.Warn("buffer drops: %d", drops)
	}
}
