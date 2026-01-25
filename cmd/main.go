package main

import (
	"log"
	"net/http"

	"myapp/internal/config"
	httpRouter "myapp/internal/http/router"
	"myapp/internal/infrastructure/firestore"
)

func main() {
	// 1. Cargar configuración
	cfg := config.Load()

	// 2. Crear cliente de Firestore
	fsClient, err := firestore.NewClient(cfg.ProjectID)
	if err != nil {
		log.Fatal(err)
	}

	// 3. Construir el router HTTP (inyectando dependencias)
	router := httpRouter.New(fsClient)

	// 4. Levantar el servidor HTTP
	log.Println("API listening on :8080")
	log.Fatal(http.ListenAndServe(":8080", router))
}
