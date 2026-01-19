package config

import (
	"os"
	"strconv"
)

type Config struct {
	ServerPort string
	DBUrl      string
	RedisAddr  string
	MinioUser  string
	MinioPass  string
	MinioEndpoint string
	MinioSecure   bool
	MinioBucket   string
	EnableImmutableTags bool
	WebhookURL string
	JWTSecret  string
	
	// Email
	SMTPHost string
	SMTPPort string
	SMTPUser string
	SMTPPass string
	SMTPFrom string

	// Cost Intelligence
	EnableCostIntelligence bool
	StorageCostPerGBMonth  float64
	BandwidthCostPerGB     float64

	// Policy
	PolicyEnvironment string
}

func Load() *Config {
	return &Config{
		ServerPort: getEnv("SERVER_PORT", ":5000"),
		DBUrl:      getEnv("DATABASE_URL", "postgres://registryx:password@localhost:5432/registryx?sslmode=disable"),
		RedisAddr:  getEnv("REDIS_ADDR", "localhost:6379"),
		MinioUser:  getEnv("MINIO_ROOT_USER", "minioadmin"),
		MinioPass:  getEnv("MINIO_ROOT_PASSWORD", "minioadmin"),
		MinioEndpoint: getEnv("MINIO_ENDPOINT", "localhost:9000"),
		MinioSecure:   getEnv("MINIO_SECURE", "false") == "true",
		MinioBucket:   getEnv("S3_BUCKET", "registryx-data"),
		EnableImmutableTags: getEnv("ENABLE_IMMUTABLE_TAGS", "false") == "true",
		PolicyEnvironment:   getEnv("POLICY_ENVIRONMENT", "dev"),
		WebhookURL: getEnv("WEBHOOK_URL", ""),
		JWTSecret:  getEnv("JWT_SECRET", "dev-secret-key-change-me"),
		
		// Email
		SMTPHost: getEnv("SMTP_HOST", ""),
		SMTPPort: getEnv("SMTP_PORT", "587"),
		SMTPUser: getEnv("SMTP_USER", ""),
		SMTPPass: getEnv("SMTP_PASS", ""),
		SMTPFrom: getEnv("SMTP_FROM", "noreply@registryx.io"),

		// Cost Defaults (AWS S3 US-East-1)
		EnableCostIntelligence: getEnv("ENABLE_COST_INTELLIGENCE", "true") == "true",
		StorageCostPerGBMonth: getEnvFloat("STORAGE_COST_PER_GB_MONTH", 0.023),
		BandwidthCostPerGB:    getEnvFloat("BANDWIDTH_COST_PER_GB", 0.09),
	}
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

func getEnvFloat(key string, fallback float64) float64 {
	if value, ok := os.LookupEnv(key); ok {
		if f, err := strconv.ParseFloat(value, 64); err == nil {
			return f
		}
	}
	return fallback
}
