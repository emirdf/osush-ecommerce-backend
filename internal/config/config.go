package config

import (
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	DBHost                string
	DBPort                string
	DBUser                string
	DBPassword            string
	DBName                string
	ServerPort            string
	RequestTimeoutSeconds int
	Env                   string
	JWTKey                string
	SecretKey             string
	MinioEndpoint         string
	MinioURL              string
	MinioAccessKey        string
	MinioSecretKey        string
	MinioBucket           string
	MinioUseSSl           bool
}

func Load(env string) *Config {
	envFile := ".env." + env

	if err := godotenv.Load(envFile); err != nil {
		log.Println("No .env file found, using environment variables")
	}

	timeout, err := strconv.Atoi(getEnv("REQUEST_TIMEOUT_SECONDS", "30"))
	if err != nil {
		log.Fatalf("Invalid REQUEST_TIMEOUT_SECONDS: %v", err)
	}

	minioUseSSL, err := strconv.ParseBool(os.Getenv("MINIO_USE_SSL"))
	if err != nil {
		log.Fatalf("Invalid MINIO_USE_SSL: %v", err)
	}

	cfg := &Config{
		DBHost:                getEnv("DB_HOST", "localhost"),
		DBPort:                getEnv("DB_PORT", "5432"),
		DBUser:                getEnv("DB_USER", "postgres"),
		DBPassword:            getEnv("DB_PASSWORD", "osush-2026"),
		DBName:                getEnv("DB_NAME", "osush"),
		ServerPort:            getEnv("SERVER_PORT", "8081"),
		RequestTimeoutSeconds: timeout,
		Env:                   getEnv("ENV", "development"),
		JWTKey:                getEnv("JWT_KEY", "eyJhbGciI2W8NvugFqHRY"),
		SecretKey:             getEnv("SECRET_KEY", "eyJhwerweirhweoirewiurhiI2W8NvugFqHRY"),
		MinioEndpoint:         getEnv("MINIO_ENDPOINT", "localhost:9000"),
		MinioURL:              getEnv("MINIO_URL", "http://localhost:9000"),
		MinioAccessKey:        getEnv("MINIO_ACCESS_KEY", "minioadmin"),
		MinioSecretKey:        getEnv("MINIO_SECRET_KEY", "minioadmin"),
		MinioBucket:           getEnv("MINIO_BUCKET", "osush"),
		MinioUseSSl:           minioUseSSL,
	}

	log.Println(cfg.DBHost, cfg.DBPort, cfg.DBUser, cfg.DBPassword, cfg.DBName)

	return cfg
}

func getEnv(key, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	return value
}

func (c *Config) MinioBaseURL() string {
	return fmt.Sprintf("%s/%s", c.MinioURL, c.MinioBucket)
}
