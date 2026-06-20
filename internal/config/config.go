package config

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/hrodrig/kiko/internal/log"
	"github.com/spf13/viper"
)

type Config struct {
	Listen       string       `mapstructure:"listen"`
	PublicURL    string       `mapstructure:"public_url"`
	LogLevel     string       `mapstructure:"log_level"`
	Database     DatabaseCfg  `mapstructure:"database"`
	Buffer       BufferCfg    `mapstructure:"buffer"`
	RateLimit    RateLimitCfg `mapstructure:"rate_limit"`
	Filter       FilterCfg    `mapstructure:"filter"`
	API          APICfg       `mapstructure:"api"`
	AllowedHosts []string     `mapstructure:"allowed_hosts"`
	Visitor      VisitorCfg   `mapstructure:"visitor"`
	Log          *log.Logger
}

type DatabaseCfg struct {
	Driver   string `mapstructure:"driver"`
	Path     string `mapstructure:"path"`
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	User     string `mapstructure:"user"`
	Password string `mapstructure:"password"`
	DBName   string `mapstructure:"dbname"`
	SSLMode  string `mapstructure:"sslmode"`

	// DSN overrides all connection fields when set.
	DSN string `mapstructure:"dsn"`
}

// NormalizedDriver returns the canonical driver name (sqlite, postgres, mysql).
func (d DatabaseCfg) NormalizedDriver() string {
	switch strings.ToLower(strings.TrimSpace(d.Driver)) {
	case "", "sqlite", "sqlite3":
		return "sqlite"
	case "postgres", "postgresql", "pg":
		return "postgres"
	case "mysql", "mariadb":
		return "mysql"
	default:
		return strings.ToLower(strings.TrimSpace(d.Driver))
	}
}

func (d DatabaseCfg) DSNString() string {
	if d.DSN != "" {
		return d.DSN
	}
	return fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=%s",
		d.User, d.Password, d.Host, d.Port, d.DBName, d.SSLMode)
}

func (d DatabaseCfg) MySQLDSN() string {
	if d.DSN != "" {
		return d.DSN
	}
	port := d.Port
	if port == 0 {
		port = 3306
	}
	user := url.UserPassword(d.User, d.Password)
	return fmt.Sprintf("%s@tcp(%s:%d)/%s?parseTime=true&charset=utf8mb4",
		user.String(), d.Host, port, d.DBName)
}

type BufferCfg struct {
	FlushInterval int `mapstructure:"flush_interval"` // seconds
	Capacity      int `mapstructure:"capacity"`       // max hits buffered before flush
}

type RateLimitCfg struct {
	Enabled            bool `mapstructure:"enabled"`
	RequestsPerSec     int  `mapstructure:"requests_per_sec"`
	Burst              int  `mapstructure:"burst"`
	HostRequestsPerSec int  `mapstructure:"host_requests_per_sec"`
	HostBurst          int  `mapstructure:"host_burst"`
}

type FilterCfg struct {
	TrustProxy         bool     `mapstructure:"trust_proxy"`
	BlockDatacenterIPs bool     `mapstructure:"block_datacenter_ips"`
	DatacenterCIDRs    []string `mapstructure:"datacenter_cidrs"`
	IgnoreIPs          []string `mapstructure:"ignore_ips"`
}

type APICfg struct {
	Key       string          `mapstructure:"key"`
	RateLimit APICfgRateLimit `mapstructure:"rate_limit"`
}

type APICfgRateLimit struct {
	Enabled        bool `mapstructure:"enabled"`
	RequestsPerSec int  `mapstructure:"requests_per_sec"`
	Burst          int  `mapstructure:"burst"`
}

type VisitorCfg struct {
	// Salt for daily visitor_hash (SHA-256). Set in production via config or KIKO_VISITOR_SALT.
	Salt string `mapstructure:"salt"`
}

func Load(path string) (*Config, error) {
	v := viper.New()
	v.SetConfigName("kiko")
	v.AddConfigPath(".")
	v.AddConfigPath("/etc/kiko/")
	v.SetEnvPrefix("KIKO")
	v.AutomaticEnv()

	v.BindEnv("database.driver", "KIKO_DATABASE_DRIVER")
	v.BindEnv("database.path", "KIKO_DATABASE_PATH")
	v.BindEnv("database.host", "KIKO_DATABASE_HOST")
	v.BindEnv("database.port", "KIKO_DATABASE_PORT")
	v.BindEnv("database.user", "KIKO_DATABASE_USER")
	v.BindEnv("database.password", "KIKO_DATABASE_PASSWORD")
	v.BindEnv("database.dbname", "KIKO_DATABASE_DBNAME")
	v.BindEnv("database.sslmode", "KIKO_DATABASE_SSLMODE")
	v.BindEnv("database.dsn", "KIKO_DATABASE_DSN")
	v.BindEnv("visitor.salt", "KIKO_VISITOR_SALT")
	v.BindEnv("rate_limit.enabled", "KIKO_RATE_LIMIT_ENABLED")
	v.BindEnv("api.key", "KIKO_API_KEY")
	v.BindEnv("api.rate_limit.enabled", "KIKO_API_RATE_LIMIT_ENABLED")
	v.BindEnv("filter.trust_proxy", "KIKO_FILTER_TRUST_PROXY")
	v.BindEnv("filter.block_datacenter_ips", "KIKO_FILTER_BLOCK_DATACENTER_IPS")

	v.SetDefault("listen", ":8080")
	v.SetDefault("public_url", "http://localhost:8080")
	v.SetDefault("log_level", "info")
	v.SetDefault("database.driver", "sqlite")
	v.SetDefault("database.path", "./data/kiko.db")
	v.SetDefault("database.host", "localhost")
	v.SetDefault("database.port", 5432)
	v.SetDefault("database.user", "kiko")
	v.SetDefault("database.password", "")
	v.SetDefault("database.dbname", "kiko")
	v.SetDefault("database.sslmode", "disable")
	v.SetDefault("buffer.flush_interval", 10)
	v.SetDefault("buffer.capacity", 4096)
	v.SetDefault("rate_limit.enabled", true)
	v.SetDefault("rate_limit.requests_per_sec", 100)
	v.SetDefault("rate_limit.burst", 200)
	v.SetDefault("rate_limit.host_requests_per_sec", 50)
	v.SetDefault("rate_limit.host_burst", 100)
	v.SetDefault("filter.trust_proxy", true)
	v.SetDefault("filter.block_datacenter_ips", false)
	v.SetDefault("api.rate_limit.enabled", true)
	v.SetDefault("api.rate_limit.requests_per_sec", 30)
	v.SetDefault("api.rate_limit.burst", 60)

	if path != "" {
		v.SetConfigFile(path)
	}

	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, err
		}
	}

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, err
	}

	if err := cfg.validate(); err != nil {
		return nil, err
	}

	lvl, err := log.ParseLevel(cfg.LogLevel)
	if err != nil {
		return nil, fmt.Errorf("config: %w", err)
	}
	cfg.Log = log.New(nil, lvl)

	return &cfg, nil
}

func (c *Config) validate() error {
	if c.Buffer.FlushInterval <= 0 {
		return fmt.Errorf("config: buffer.flush_interval must be > 0")
	}
	if c.Buffer.Capacity <= 0 {
		return fmt.Errorf("config: buffer.capacity must be > 0")
	}
	if c.RateLimit.Enabled && c.RateLimit.RequestsPerSec <= 0 {
		return fmt.Errorf("config: rate_limit.requests_per_sec must be > 0 when rate limiting is enabled")
	}
	if c.RateLimit.Burst <= 0 {
		return fmt.Errorf("config: rate_limit.burst must be > 0")
	}
	if c.API.RateLimit.Enabled && c.API.RateLimit.RequestsPerSec <= 0 {
		return fmt.Errorf("config: api.rate_limit.requests_per_sec must be > 0 when enabled")
	}
	if c.API.RateLimit.Burst <= 0 {
		c.API.RateLimit.Burst = 60
	}
	switch c.Database.NormalizedDriver() {
	case "sqlite", "postgres", "mysql":
	default:
		return fmt.Errorf("config: database.driver %q unsupported (want sqlite, postgres, mysql)", c.Database.Driver)
	}
	if c.Database.NormalizedDriver() == "mysql" && c.Database.DSN == "" && c.Database.Port == 5432 {
		c.Database.Port = 3306
	}
	return nil
}
