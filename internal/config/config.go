package config

import (
	"strings"

	"github.com/spf13/viper"
)

type Config struct {
	App      AppConfig
	Database DatabaseConfig
	RabbitMQ RabbitMQConfig
	Auth     AuthConfig
	Log      LogConfig
}

type AppConfig struct {
	Env  string
	Port int
	Host string
}

type DatabaseConfig struct {
	URL      string
	MaxConns int32
	MinConns int32
}

type RabbitMQConfig struct {
	URL string
}

type AuthConfig struct {
	JWTSecret      string
	InternalSecret string
}

type LogConfig struct {
	Level string
}

func Load() (*Config, error) {
	viper.SetConfigFile(".env")
	viper.SetConfigType("env")
	viper.ReadInConfig()

	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()

	viper.SetDefault("APP_PORT", 8082)
	viper.SetDefault("APP_HOST", "0.0.0.0")
	viper.SetDefault("APP_ENV", "development")
	viper.SetDefault("DATABASE_MAX_CONNS", 25)
	viper.SetDefault("DATABASE_MIN_CONNS", 5)
	viper.SetDefault("RABBITMQ_URL", "amqp://guest:guest@localhost:5672/")
	viper.SetDefault("LOG_LEVEL", "info")

	return &Config{
		App: AppConfig{
			Env:  viper.GetString("APP_ENV"),
			Port: viper.GetInt("APP_PORT"),
			Host: viper.GetString("APP_HOST"),
		},
		Database: DatabaseConfig{
			URL:      viper.GetString("DATABASE_URL"),
			MaxConns: int32(viper.GetInt("DATABASE_MAX_CONNS")),
			MinConns: int32(viper.GetInt("DATABASE_MIN_CONNS")),
		},
		RabbitMQ: RabbitMQConfig{
			URL: viper.GetString("RABBITMQ_URL"),
		},
		Auth: AuthConfig{
			JWTSecret:      viper.GetString("JWT_SECRET"),
			InternalSecret: viper.GetString("INTERNAL_SECRET"),
		},
		Log: LogConfig{
			Level: viper.GetString("LOG_LEVEL"),
		},
	}, nil
}
