package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

type Config struct {
	Server      ServerConfig
	Swagger     SwaggerConfig
	Postgres    PostgresConfig
	MongoDB     MongoDBConfig
	Redis       RedisConfig
	JWT         JWTConfig
	Options     map[string][]string `json:"options"`
	Environment string
}

type ServerConfig struct {
	Port string
}

type SwaggerConfig struct {
	Port string
}

type PostgresConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	Database string
	SSLMode  string
	Options  map[string][]string `json:"options"`
}

type MongoDBConfig struct {
	URI      string
	Database string
}

type RedisConfig struct {
	Addr     string
	Password string
	DB       int
}

type JWTConfig struct {
	Secret     string
	Expiration int // in hours
}

var cnf Config

func GetConfig() Config {
	return cnf
}

func Load() *Config {
	cnf = Config{
		Environment: getEnv("ENVIRONMENT", "development"),
		Server: ServerConfig{
			Port: getEnv("SERVER_PORT", "8080"),
		},
		Swagger: SwaggerConfig{
			Port: getEnv("SWAGGER_PORT", "8081"),
		},
		Postgres: PostgresConfig{
			Host:     getEnv("POSTGRES_HOST", "localhost"),
			Port:     getEnvAsInt("POSTGRES_PORT", 5436),
			User:     getEnv("POSTGRES_USER", "root"),
			Password: getEnv("POSTGRES_PASSWORD", "secret"),
			Database: getEnv("POSTGRES_DB", "ride_engine"),
			SSLMode:  getEnv("POSTGRES_SSLMODE", "disable"),
			//Options:  viperOrEnvStringMapSlice("POSTGRES_OPTIONS", "sslmode=disable"),
			Options: map[string][]string{
				"sslmode": []string{"disable"},
			},
		},
		MongoDB: MongoDBConfig{
			URI:      getEnv("MONGODB_URI", "mongodb://root:secret@localhost:27016/?authSource=admin"),
			Database: getEnv("MONGODB_DATABASE", "ride_engine"),
		},
		Redis: RedisConfig{
			Addr:     getRedisAddr(),
			Password: getEnv("REDIS_PASSWORD", ""),
			DB:       getEnvAsInt("REDIS_DB", 0),
		},
		JWT: JWTConfig{
			Secret:     getEnv("JWT_SECRET", "your-secret-key-change-in-production"),
			Expiration: getJWTExpiration(),
		},
	}

	if cnf.Environment == "development" {
		cnf.JWT.Expiration = 10000 // 10000 second expiry
	}
	return &cnf
}

func (c *PostgresConfig) DSN() string {
	return fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		c.Host, c.Port, c.User, c.Password, c.Database, c.SSLMode,
	)
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
	valueStr := getEnv(key, "")
	if value, err := strconv.Atoi(valueStr); err == nil {
		return value
	}
	return defaultValue
}

func getRedisAddr() string {
	if addr := os.Getenv("REDIS_ADDR"); addr != "" {
		return addr
	}

	host := getEnv("REDIS_HOST", "localhost")
	port := getEnv("REDIS_PORT", "6379")
	return fmt.Sprintf("%s:%s", host, port)
}

func getJWTExpiration() int {
	if expStr := os.Getenv("JWT_EXPIRATION"); expStr != "" {
		if duration, err := time.ParseDuration(expStr); err == nil {
			return int(duration.Hours())
		}
		if hours, err := strconv.Atoi(expStr); err == nil {
			return hours
		}
	}

	if hours := getEnvAsInt("JWT_EXPIRATION_HOURS", 0); hours > 0 {
		return hours
	}

	return 24
}
