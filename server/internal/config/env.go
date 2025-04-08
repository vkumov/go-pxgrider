package config

import (
	"os"

	"github.com/joho/godotenv"
)

func loadEnv() {
	env := os.Getenv("ENV")
	if env == "" {
		env = "production"
	}

	godotenv.Load(".env." + env + ".local")
	if env != "test" {
		godotenv.Load(".env.local")
	}
	godotenv.Load(".env." + env)
	godotenv.Load() // The Original .env
}
