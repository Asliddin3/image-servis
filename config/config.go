package config

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
	"github.com/spf13/cast"
)

type Config struct {
	Environment      string
	LogLevel         string
	PostgresHost     string
	PostgresPort     string
	PostgresDatabase string
	PostgresUser     string
	PostgresPassword string
	ImageServiceHost string
	ImageServicePort string
	ImagePath        string
	PGXPoolMax       int
}

func LoadConfig() *Config {
	_ = godotenv.Load(".env")
	c := &Config{}
	c.Environment = cast.ToString(getOrReturnDefault("ENVIRONMENT", "develop"))
	c.LogLevel = cast.ToString(getOrReturnDefault("LOG_LEVEL", "debug"))
	c.PostgresHost = cast.ToString(getOrReturnDefault("POSTGRES_HOST", "localhost"))
	c.PostgresPort = cast.ToString(getOrReturnDefault("POSTGRES_PORT", 5432))
	c.PostgresDatabase = cast.ToString(getOrReturnDefault("POSTGRES_DATABASE", "imagedb"))
	c.PostgresUser = cast.ToString(getOrReturnDefault("POSTGRES_USER", "postgres"))
	c.PostgresPassword = cast.ToString(getOrReturnDefault("POSTGRES_PASSWORD", "compos1995"))
	c.ImageServiceHost = cast.ToString(getOrReturnDefault("IMAGE_SERVICE_HOST", "localhost"))
	c.ImageServicePort = cast.ToString(getOrReturnDefault("IMAGE_SERVICE_PORT", "7000"))
	c.ImagePath = cast.ToString(getOrReturnDefault("IMAGE_PATH", "../../../img"))
	c.PGXPoolMax = cast.ToInt(getOrReturnDefault("PGX_POOL_MAX", 2))

	return c
}

func getOrReturnDefault(key string, defaultValue interface{}) interface{} {
	_, exists := os.LookupEnv(key)
	fmt.Println(exists)
	if exists {
		return os.Getenv(key)
	}
	return defaultValue
}
