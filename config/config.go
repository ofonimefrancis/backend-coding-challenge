package config

import (
	"fmt"
	"time"

	"github.com/joeshaw/envdecode"
	"github.com/joho/godotenv"
)

// Configuration struct to hold all the configuration for the application
type Configuration struct {
	Server   ServerConfig
	Database Postgres
	JWT      JWTConfig
	Redis    RedisConfig
	AppName  string `env:"APP_NAME,default=[thermondo-backend]: "`
}

type ServerConfig struct {
	Port                string        `env:"SERVER_PORT,default=8080"`
	Host                string        `env:"SERVER_HOST,default=localhost"`
	ReadTimeout         time.Duration `env:"SERVER_READ_TIMEOUT,default=10s"`
	WriteTimeout        time.Duration `env:"SERVER_WRITE_TIMEOUT,default=10s"`
	IdleTimeout         time.Duration `env:"SERVER_IDLE_TIMEOUT,default=10s"`
	ShutdownTimeout     time.Duration `env:"SERVER_SHUTDOWN_TIMEOUT,default=10s"`
	ShutdownGracePeriod time.Duration `env:"SERVER_SHUTDOWN_GRACE_PERIOD,default=10s"`
}

type Postgres struct {
	DSN          string `env:"POSTGRESQL_DSN,default=host=localhost dbname=postgres user=postgres sslmode=disable"`
	MaxIdleConns int    `env:"POSTGRES_MAX_IDLE_CONNECTIONS,default=20"`
	MaxOpenConns int    `env:"POSTGRES_MAX_OPEN_CONNECTIONS,default=20"`
	HealthCheck  bool   `env:"POSTGRES_HEALTH_CHECK,default=false"`
}

type JWTConfig struct {
	Secret string        `env:"JWT_SECRET,default=secret"`
	Expiry time.Duration `env:"JWT_EXPIRY,default=1h"`
}

type RedisConfig struct {
	Host     string `env:"REDIS_HOST,default=localhost"`
	Port     string `env:"REDIS_PORT,default=6379"`
	Password string `env:"REDIS_PASSWORD,default=password"`
	DB       int    `env:"REDIS_DB,default=0"`
}

// LoadConfig loads the configuration from the environment variables
func LoadConfig() (Configuration, error) {

	if err := godotenv.Load(); err != nil {
		fmt.Println("Warning: .env file not found, using environment variables")
	}

	var conf Configuration
	if err := envdecode.Decode(&conf); err != nil {
		return Configuration{}, err
	}
	return conf, nil
}
