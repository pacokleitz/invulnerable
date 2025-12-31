package config

import (
	"fmt"
	"os"
)

// Config holds all application configuration
type Config struct {
	Database DatabaseConfig
	S3       S3Config
	Server   ServerConfig
}

// DatabaseConfig holds database connection settings
type DatabaseConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	DBName   string
	SSLMode  string
}

// S3Config holds S3/MinIO configuration for SBOM storage
type S3Config struct {
	Endpoint  string
	Bucket    string
	Region    string
	AccessKey string
	SecretKey string
	UseSSL    bool
}

// ServerConfig holds HTTP server configuration
type ServerConfig struct {
	Port string
}

// LoadFromEnv loads configuration from environment variables
func LoadFromEnv() (*Config, error) {
	config := &Config{
		Database: DatabaseConfig{
			Host:     getEnv("DB_HOST", "localhost"),
			Port:     getEnv("DB_PORT", "5432"),
			User:     getEnv("DB_USER", "postgres"),
			Password: getEnv("DB_PASSWORD", "postgres"),
			DBName:   getEnv("DB_NAME", "invulnerable"),
			SSLMode:  getEnv("DB_SSLMODE", "disable"),
		},
		S3: S3Config{
			Endpoint:  getEnv("SBOM_S3_ENDPOINT", ""),
			Bucket:    getEnv("SBOM_S3_BUCKET", "invulnerable"),
			Region:    getEnv("SBOM_S3_REGION", "us-east-1"),
			AccessKey: getEnv("SBOM_S3_ACCESS_KEY", ""),
			SecretKey: getEnv("SBOM_S3_SECRET_KEY", ""),
			UseSSL:    getEnv("SBOM_S3_USE_SSL", "true") == "true",
		},
		Server: ServerConfig{
			Port: getEnv("PORT", "8080"),
		},
	}

	// Validate required S3 settings
	if config.S3.Endpoint == "" {
		return nil, fmt.Errorf("SBOM_S3_ENDPOINT is required")
	}
	if config.S3.AccessKey == "" {
		return nil, fmt.Errorf("SBOM_S3_ACCESS_KEY is required")
	}
	if config.S3.SecretKey == "" {
		return nil, fmt.Errorf("SBOM_S3_SECRET_KEY is required")
	}

	return config, nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// GetDatabaseDSN returns the PostgreSQL connection string
func (c *Config) GetDatabaseDSN() string {
	return fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		c.Database.Host,
		c.Database.Port,
		c.Database.User,
		c.Database.Password,
		c.Database.DBName,
		c.Database.SSLMode,
	)
}
