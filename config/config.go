package config

import (
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	ServerAddress string
	PostgresConn  string
	GinMode       string
}

func LoadConfig() (*Config, error) {
	err := godotenv.Load()
	if err != nil {
		//log.Fatal("Error loading .env file")
	}

	serverAddress := os.Getenv("SERVER_ADDRESS")
	if serverAddress == "" {
		serverAddress = "0.0.0.0:8080"
	}

	ginMode := os.Getenv("GIN_MODE")
	if ginMode == "" {
		ginMode = "release"
	}

	config := &Config{
		ServerAddress: serverAddress,
		PostgresConn:  os.Getenv("POSTGRES_CONN"),
		GinMode:       ginMode,
	}

	return config, nil
}
