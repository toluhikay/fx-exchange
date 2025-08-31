package config

import (
	"os"
	"time"

	"github.com/toluhikay/fx-exchange/pkg/jwt"
)

type Config struct {
	Port       string
	DbHost     string
	DbPort     string
	DbName     string
	DbUser     string
	DbPassword string
	Auth       jwt.Auth
}

func LoadConfig() Config {

	return Config{
		Port:       getOrDefaultEnv("PORT", "8080"),
		DbHost:     getOrDefaultEnv("DB_HOST", "localhost"),
		DbPort:     getOrDefaultEnv("DB_PORT", "5433"),
		DbName:     getOrDefaultEnv("DB_NAME", "fx-exchange"),
		DbUser:     getOrDefaultEnv("DB_USER", "fx-exchange"),
		DbPassword: getOrDefaultEnv("DB_PASSWORD", "fx-exchange"),
		Auth: jwt.Auth{
			Issuer:               getOrDefaultEnv("JWT_ISSUER", "fx-exchange.com"),
			Audience:             getOrDefaultEnv("JWT_AUDIENCE", "fx-exchange"),
			Secret:               getOrDefaultEnv("JWT_SECRET", "fx-exchange-jwt-secret"),
			TokenExpireAt:        time.Minute * 15,
			RefreshTokenExpireAt: time.Hour * 4,
		},
	}
}

func getOrDefaultEnv(key, fallback string) string {
	val, ok := os.LookupEnv(key)
	if !ok {
		return fallback
	}

	return val
}
