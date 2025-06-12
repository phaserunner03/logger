// main.go
package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/phaserunner03/logging/configs"

	"github.com/phaserunner03/logging/internal/router"
)



func main() {


	config, err := configs.LoadConfig()
	if err != nil {
		log.Fatalf("Error loading configuration: %v", err)
	}

	port := config.Services.Port
	r:= router.Router()
	if err != nil {
		log.Fatalf("Error loading configuration: %v", err)
	}
	log.Printf("Starting server on port %d", port)
	if err := http.ListenAndServe(":"+fmt.Sprintf("%d", port), r); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}

}
