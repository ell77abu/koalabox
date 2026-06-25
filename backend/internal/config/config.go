package config

import (
	"fmt"
	"strings"

	"github.com/spf13/viper"
)

type Config struct {
	Server   ServerConfig   `mapstructure:"server"`
	Database DatabaseConfig `mapstructure:"database"`
	JWT      JWTConfig      `mapstructure:"jwt"`
}

type ServerConfig struct {
	Host string `mapstructure:"host"`
	Port int    `mapstructure:"port"`
}

type DatabaseConfig struct {
	DSN                    string `mapstructure:"dsn"`
	MaxIdleConns           int    `mapstructure:"max_idle_conns"`
	MaxOpenConns           int    `mapstructure:"max_open_conns"`
	ConnMaxLifetimeMinutes int    `mapstructure:"conn_max_lifetime_minutes"`
}

type JWTConfig struct {
	Secret     string `mapstructure:"secret"`
	ExpireHour int    `mapstructure:"expire_hour"`
}

func Load() (*Config, error) {
	v := viper.New()

	v.SetConfigName("config")
	v.SetConfigType("yaml")
	v.AddConfigPath(".")
	v.AddConfigPath("./configs")

	v.SetEnvPrefix("APP")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	setDefaults(v)

	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("read config file: %w", err)
		}
	}

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("unmarshal config: %w", err)
	}

	if cfg.Database.DSN == "" {
		return nil, fmt.Errorf("database.dsn is required")
	}
	if cfg.Database.MaxIdleConns < 0 {
		return nil, fmt.Errorf("database.max_idle_conns must be non-negative")
	}
	if cfg.Database.MaxOpenConns <= 0 {
		return nil, fmt.Errorf("database.max_open_conns must be greater than zero")
	}
	if cfg.Database.ConnMaxLifetimeMinutes < 0 {
		return nil, fmt.Errorf("database.conn_max_lifetime_minutes must be non-negative")
	}
	if cfg.JWT.Secret == "" {
		return nil, fmt.Errorf("jwt.secret is required")
	}
	if cfg.JWT.ExpireHour <= 0 {
		return nil, fmt.Errorf("jwt.expire_hour must be greater than zero")
	}

	return &cfg, nil
}

func setDefaults(v *viper.Viper) {
	v.SetDefault("server.host", "0.0.0.0")
	v.SetDefault("server.port", 8080)

	v.SetDefault("database.max_idle_conns", 10)
	v.SetDefault("database.max_open_conns", 50)
	v.SetDefault("database.conn_max_lifetime_minutes", 60)

	v.SetDefault("jwt.expire_hour", 72)
}
