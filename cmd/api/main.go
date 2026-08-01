package main

import (
	"context"
	"log"

	"github.com/joho/godotenv"
	"github.com/snsilvam/kaizensnsilvam-backend/internal/config"
	"github.com/snsilvam/kaizensnsilvam-backend/internal/family"
	router "github.com/snsilvam/kaizensnsilvam-backend/internal/http"
	"github.com/snsilvam/kaizensnsilvam-backend/internal/http/handlers"
	fs "github.com/snsilvam/kaizensnsilvam-backend/internal/infrastructure/firestore"
	"github.com/snsilvam/kaizensnsilvam-backend/internal/user"
)

func main() {
	// Carga variables desde .env si existe (en local). Se prueba el
	// directorio actual y la raíz del proyecto, porque `go run` usa como
	// working dir la carpeta desde donde se ejecuta (p.ej. cmd/api).
	// En producción las variables vienen del entorno y no hay archivo .env.
	loadDotEnv(".env", "../../.env")

	cfg := config.Load()

	ctx := context.Background()

	// Infraestructura: cliente de Firestore.
	fsClient, err := fs.NewClient(ctx, cfg.ProjectID)
	if err != nil {
		log.Fatalf("firestore: %v", err)
	}
	defer fsClient.Close()

	// Composition root: cableado de cada feature.
	familyRepo := fs.NewFamilyRepository(fsClient)
	familySvc := family.NewService(familyRepo)
	familyHandler := handlers.NewFamilyHandler(familySvc)

	userRepo := fs.NewUserRepository(fsClient)
	userSvc := user.NewService(userRepo)
	userHandler := handlers.NewUserHandler(userSvc)

	r := router.New(familyHandler, userHandler)

	log.Println("API listening on :" + cfg.Port)
	if err := r.Run(":" + cfg.Port); err != nil {
		log.Fatal(err)
	}
}

// loadDotEnv carga el primer archivo .env que exista de la lista.
// godotenv.Load corta en el primer error, así que hay que probar
// cada ruta por separado.
func loadDotEnv(paths ...string) {
	for _, p := range paths {
		if err := godotenv.Load(p); err == nil {
			log.Printf("loaded env from %s", p)
			return
		}
	}
	log.Println("no .env file found, using environment variables")
}
