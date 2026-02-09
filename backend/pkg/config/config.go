package config

import (
	"log"
	"os"
	"strconv"
	"strings"
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
	
	// URLs
	BackendURL  string
	FrontendURL string
	
	// Email
	SMTPHost string
	SMTPPort string
	SMTPUser string
	SMTPPass string
	SMTPFrom string

	// Cost Intelligence
	EnableCostIntelligence bool
	CostMode               string  // "CLOUD" or "ONPREM"
	StorageCapacityTB      float64 // Total capacity in TB (for On-Prem efficiency)
	StorageCostPerGBMonth  float64
	BandwidthCostPerGB     float64

	// Policy
	PolicyEnvironment string

	// Security
	CORSAllowedOrigins string
}

func Load() *Config {
	policyEnv := getEnv("POLICY_ENVIRONMENT", "dev")
	
	// JWT Secret validation
	jwtSecret := getEnv("JWT_SECRET", "")
	if jwtSecret == "" || jwtSecret == "dev-secret-key-change-me" {
		if policyEnv == "prod" {
			log.Fatal("❌ SECURITY ERROR: JWT_SECRET must be set to a strong value in production")
		}
		log.Println("⚠️  WARNING: Using default JWT secret. Set JWT_SECRET in .env for production")
		jwtSecret = "dev-secret-key-change-me"
	}
	
	// Database URL validation
	dbUrl := getEnv("DATABASE_URL", "postgres://registryx:password@localhost:5432/registryx?sslmode=disable")
	if policyEnv == "prod" && !strings.Contains(dbUrl, "sslmode=require") && !strings.Contains(dbUrl, "sslmode=verify") {
		log.Println("⚠️  WARNING: Database SSL is not enabled in production. Consider using sslmode=require")
	}
	
	return &Config{
		ServerPort: getEnv("SERVER_PORT", ":5000"),
		DBUrl:      dbUrl,
		RedisAddr:  getEnv("REDIS_ADDR", "localhost:6379"),
		MinioUser:  getEnv("MINIO_ROOT_USER", "minioadmin"),
		MinioPass:  getEnv("MINIO_ROOT_PASSWORD", "minioadmin"),
		MinioEndpoint: getEnv("MINIO_ENDPOINT", "localhost:9000"),
		MinioSecure:   getEnv("MINIO_SECURE", "false") == "true",
		MinioBucket:   getEnv("S3_BUCKET", "registryx-data"),
		EnableImmutableTags: getEnv("ENABLE_IMMUTABLE_TAGS", "false") == "true",
		PolicyEnvironment:   policyEnv,
		WebhookURL: getEnv("WEBHOOK_URL", ""),
		JWTSecret:  jwtSecret,
		
		// URLs
		BackendURL:  getEnv("BACKEND_URL", "http://localhost:5000"),
		FrontendURL: getEnv("FRONTEND_URL", "http://localhost:5173"),
		
		// Email
		SMTPHost: getEnv("SMTP_HOST", ""),
		SMTPPort: getEnv("SMTP_PORT", "587"),
		SMTPUser: getEnv("SMTP_USER", ""),
		SMTPPass: getEnv("SMTP_PASS", ""),
		SMTPFrom: getEnv("SMTP_FROM", "noreply@registryx.io"),

		// Cost Defaults (AWS S3 US-East-1)
		EnableCostIntelligence: getEnv("ENABLE_COST_INTELLIGENCE", "true") == "true",
		CostMode:               getEnv("COST_MODE", "CLOUD"), // Default: CLOUD
		StorageCapacityTB:      getEnvFloat("STORAGE_CAPACITY_TB", 10.0), // Default 10TB
		StorageCostPerGBMonth:  getEnvFloat("STORAGE_COST_PER_GB_MONTH", 0.023),
		BandwidthCostPerGB:     getEnvFloat("BANDWIDTH_COST_PER_GB", 0.09),

		// Security
		CORSAllowedOrigins: getEnv("CORS_ALLOWED_ORIGINS", "http://localhost:5173"), // Comma separated
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
