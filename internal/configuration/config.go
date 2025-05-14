package configuration

import (
	"cloud-test/internal/backends"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

type Config struct {
	ProxyListen  string                 `yaml:"proxy_listen"`
	AdminListen  string                 `yaml:"admin_listen"`
	Loadbalancer backends.Configuration `yaml:"loadbalancer"`
	RateLimit    RateCfg                `yaml:"rate_limit"`
	PgDb         PgDbCfg                `yaml:"postgres_database"`
	Redis        RedisCfg               `yaml:"redis_database"`

	LoggerPath string `yaml:"log_path"`
}

type RedisCfg struct {
	Addr     string `yaml:"addr"`
	Password string `yaml:"password"`
	Db       int    `yaml:"db"`
}

type PgDbCfg struct {
	Host           string `yaml:"host"`
	Port           int    `yaml:"port"`
	User           string `yaml:"user"`
	Password       string `yaml:"password"`
	DbName         string `yaml:"dbname"`
	MigrationsPath string `yaml:"migrations_path"`
}

type HealthCfg struct {
	Interval time.Duration `yaml:"interval"`
}

type RateCfg struct {
	Capacity       int           `yaml:"capacity"`
	RefillInterval time.Duration `yaml:"refill_interval"`
	Enabled        bool          `yaml:"enabled"`
}

type Backend struct {
	Url         string             `yaml:"url"`
	Name        string             `yaml:"name"`
	HealthCheck BackendHealthCheck `yaml:"health_check"`
}

type BackendHealthCheck struct {
	Timeout time.Duration `yaml:"timeout"`
	Url     string        `yaml:"url"`
	Method  string        `yaml:"method"`
}

func Load(path string) (Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Config{}, err
	}
	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return Config{}, err
	}

	return cfg, nil
}
