package handler
import (
	"context"
	"encoding/json"
	"log"
	
	"net/http"
	"github.com/phaserunner03/logging/configs"
    "github.com/phaserunner03/logging/internal/logs"
	
)

func ProcessLogsHandler(w http.ResponseWriter, r *http.Request) {
	
	ctx := context.Background()

	config, err := configs.LoadConfig()
	if err != nil {
		http.Error(w, "Failed to load configuration", http.StatusInternalServerError)
		log.Printf("Error loading configuration: %v", err)
		return
	}

	services := config.Services.Name
	startDate := "2023-01-01T00:00:00Z" 
	endDate := "2023-12-31T23:59:59Z"

	if err:= logs.ProcessLogs(ctx,services,startDate,endDate); err!=nil{
		log.Printf("Error processing logs: %v", err)
        http.Error(w, "Failed to process logs", http.StatusInternalServerError)
        return
	}

	response := map[string]string{"status": "Logs processed successfully"}
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		log.Printf("Error encoding response: %v", err)
		return
	}
	log.Printf("Successfully processed logs for services: %v", services)
	json.NewEncoder(w).Encode(response)
}