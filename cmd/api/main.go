package main

import (
	"log"

	"github.com/snsilvam/kaizensnsilvam-backend/internal/config"
	"github.com/snsilvam/kaizensnsilvam-backend/internal/http/router"
)

func main() {
	cfg := config.Load()

	r := router.New()

	log.Println("API listening on :" + cfg.Port)
	if err := r.Run(":" + cfg.Port); err != nil {
		log.Fatal(err)
	}
}
