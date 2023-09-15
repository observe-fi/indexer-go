package app

import (
	"github.com/joho/godotenv"
	"log"
)

func MustLoadEnv() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	ReloadConfig()
}
