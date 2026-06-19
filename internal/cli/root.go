package cli

import (
	"net/http"

	"github.com/hrodrig/kiko/internal/config"
	"github.com/hrodrig/kiko/internal/hit"
	"github.com/hrodrig/kiko/internal/server"
	"github.com/hrodrig/kiko/internal/store"
	"github.com/hrodrig/kiko/internal/version"
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
			cmd.Printf("kiko %s (commit %s, built %s, branch %s)\n",
				version.Version, version.Commit, version.BuildDate, version.Branch)
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

	buf := hit.NewBuffer()
	st := store.NewNop()

	sv := server.New(st, buf, cfg.Log, cfg.AllowedHosts)
	http.Handle("/", sv.Handler())

	cfg.Log.Info("kiko v%s starting on %s (buffer flush=%ds, cap=%d)",
		version.Version, cfg.Listen, cfg.Buffer.FlushInterval, cfg.Buffer.Capacity)
	cfg.Log.Info("public_url: %s", cfg.PublicURL)

	return http.ListenAndServe(cfg.Listen, nil)
}
