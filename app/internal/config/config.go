package config

import (
	"flag"
	"log"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/WM1rr0rB8/librariesTest/backend/golang/logging"
	"github.com/ilyakaznacheev/cleanenv"
)

type AppConfig struct {
	ID            string `yaml:"id" env:"APP_ID"`
	Name          string `yaml:"name" env:"APP_NAME"`
	Version       string `yaml:"version" env:"APP_VERSION"`
	IsDevelopment bool   `yaml:"is_dev" env:"APP_IS_DEVELOPMENT"`
	LogLevel      string `yaml:"log_level" env:"APP_LOG_LEVEL"`
	IsLogJSON     bool   `yaml:"is_log_json" env:"APP_IS_LOG_JSON"`
	Domain        string `yaml:"domain" env:"APP_DOMAIN"`
}

type GRPCConfig struct {
	Host                string        `yaml:"host" env:"GRPC_HOST"`
	Port                int           `yaml:"port" env:"GRPC_PORT"`
	HealthCheckInterval time.Duration `yaml:"health_check_interval" env:"GRPC_HEALTH_CHECK_INTERVAL"`
}

type HTTPConfig struct {
	Host              string        `yaml:"host" env:"HTTP_HOST"`
	Port              int           `yaml:"port" env:"HTTP_PORT"`
	ReadHeaderTimeout time.Duration `yaml:"read_header_timeout" env:"HTTP_READ_HEADER_TIMEOUT"`
}

type PostgresConfig struct {
	Host       string        `yaml:"host"  env:"POSTGRES_HOST"`
	User       string        `yaml:"user" env:"POSTGRES_USER"`
	Password   string        `yaml:"password" env:"POSTGRES_PASSWORD"`
	Port       int           `yaml:"port" env:"POSTGRES_PORT"`
	Database   string        `yaml:"database" env:"POSTGRES_DATABASE"`
	MaxAttempt int           `yaml:"max_attempt"`
	MaxDelay   time.Duration `yaml:"max_delay"`
	Binary     bool          `yaml:"binary" env:"POSTGRES_BINARY"`
}

type TracingConfig struct {
	Enabled bool   `yaml:"enabled" env:"TRACING_ENABLED"`
	Host    string `yaml:"host" env:"TRACING_HOST"`
	Port    int    `yaml:"port" env:"TRACING_PORT"`
}

type MetricsConfig struct {
	Host              string        `yaml:"host" env:"METRICS_HOST"`
	Port              int           `yaml:"port" env:"METRICS_PORT"`
	ReadTimeout       time.Duration `yaml:"read_timeout"`
	WriteTimeout      time.Duration `yaml:"write_timeout"`
	ReadHeaderTimeout time.Duration `yaml:"read_header_timeout"`
	Enabled           bool          `yaml:"enabled" env:"METRICS_ENABLED"`
}

type PacksSizeConfig struct {
	PackSize []int `yaml:"pack_size" env:"PACKS_SIZE_PACK"`
}

type Config struct {
	App       AppConfig       `yaml:"app"`
	GRPC      GRPCConfig      `yaml:"grpc"`
	HTTP      HTTPConfig      `yaml:"http"`
	Postgres  PostgresConfig  `yaml:"postgres"`
	Tracing   TracingConfig   `yaml:"tracing"`
	Metrics   MetricsConfig   `yaml:"metrics"`
	PacksSize PacksSizeConfig `yaml:"packs_size"`
}

func (i *Config) LogValue() logging.Value {
	return logging.GroupValue(
		logging.Group("app",
			logging.StringAttr("id", i.App.ID),
			logging.StringAttr("name", i.App.Name),
			logging.StringAttr("version", i.App.Version),
			logging.BoolAttr("is_dev", i.App.IsDevelopment),
			logging.StringAttr("log_level", i.App.LogLevel),
			logging.BoolAttr("is_logjson", i.App.IsLogJSON),
			logging.StringAttr("domain", i.App.Domain),
		),
		logging.Group("grpc-server",
			logging.StringAttr("host", i.GRPC.Host),
			logging.IntAttr("port", i.GRPC.Port),
		),
		logging.Group("http-server",
			logging.StringAttr("host", i.HTTP.Host),
			logging.IntAttr("port", i.HTTP.Port),
		),
		logging.Group("postgres",
			logging.StringAttr("host", i.Postgres.Host),
			logging.StringAttr("user", i.Postgres.User),
			logging.StringAttr("password", strconv.Itoa(len(i.Postgres.Password))),
			logging.IntAttr("port", i.Postgres.Port),
			logging.StringAttr("database", i.Postgres.Database),
			logging.IntAttr("max_attempt", i.Postgres.MaxAttempt),
			logging.StringAttr("max_delay", i.Postgres.MaxDelay.String()),
			logging.BoolAttr("binary", i.Postgres.Binary),
		),
		logging.Group("tracing",
			logging.BoolAttr("enabled", i.Tracing.Enabled),
			logging.StringAttr("host", i.Tracing.Host),
			logging.IntAttr("port", i.Tracing.Port),
		),
		logging.Group("metrics",
			logging.StringAttr("host", i.Metrics.Host),
			logging.IntAttr("port", i.Metrics.Port),
			logging.BoolAttr("enabled", i.Metrics.Enabled),
		),
	)
}

const (
	FlagConfigPathName = "config"
	EnvConfigPathName  = "CONFIG_PATH"
)

var (
	configPath string
	instance   Config
	once       sync.Once
)

func GetConfig() *Config {
	once.Do(func() {
		flag.StringVar(
			&configPath,
			FlagConfigPathName,
			"",
			"this is application configuration file",
		)
		flag.Parse()

		if path, ok := os.LookupEnv(EnvConfigPathName); ok {
			configPath = path
		}

		log.Printf("config initializing from: %s", configPath)

		instance = Config{}

		if err := cleanenv.ReadConfig(configPath, &instance); err != nil {
			help, _ := cleanenv.GetDescription(&instance, nil)
			log.Println(help)
			log.Fatal(err)
		}

		log.Println("configuration loaded")
	})

	return &instance
}
