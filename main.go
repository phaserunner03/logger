// main.go
package main

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/phaserunner03/logging/configs"
	"github.com/phaserunner03/logging/internal/analysis"
	"github.com/phaserunner03/logging/internal/bigquery"
	"github.com/phaserunner03/logging/internal/logs"
	"github.com/phaserunner03/logging/internal/router"
)


func processLogs(ctx context.Context, services []string, startDate, endDate string) error {
	// Fetch logs from Cloud Logging
	entries, err := logs.FetchLogs(ctx, services, startDate, endDate)
	if err != nil {
		return fmt.Errorf("failed to fetch logs: %v", err)
	}

	if len(entries) == 0 {
		log.Println("No log entries to process")
		return nil
	}

	// Pre-allocate slice with exact capacity needed
	bqRows := make([]configs.BQLogRow, 0, len(entries))
	var conversionErrors int

	// Convert log entries to BigQuery rows in batch
	for _, entry := range entries {
		row, err := logs.ConvertToBQRow(entry)
		if err != nil {
			log.Printf("Warning: Failed to convert log entry: %v", err)
			conversionErrors++
			continue
		}
		row.ServiceName = entry.GetResource().GetLabels()["service_name"] // Add service name to row
		bqRows = append(bqRows, row)
	}

	if len(bqRows) == 0 {
		return fmt.Errorf("all %d log entries failed to convert", len(entries))
	}
	//if error in bqRows after analysing all entries, publish message to Pub/Sub
	if err := analysis.HandleError(ctx, bqRows); err != nil {
		return fmt.Errorf("failed to handle errors: %v", err)
	}
	// Insert all rows into BigQuery in a single batch
	if err := bigquery.InsertLogs(ctx, bqRows); err != nil {
		return fmt.Errorf("failed to insert logs into BigQuery: %v", err)
	}

	log.Printf("Successfully processed %d log entries (%d conversion errors)", len(bqRows), conversionErrors)
	return nil
}

func main() {


	// ctx := context.Background()
	// config, err := configs.LoadConfig()
	// if err != nil {
	// 	log.Fatalf("Error loading configuration: %v", err)
	// }

	// fmt.Println(config.Services.Name)

	// services := config.Services.Name    
	// startDate := "2025-06-01T00:00:00Z" 
	// endDate := "2025-06-05T23:59:59Z"   

	// if err := processLogs(ctx, services, startDate, endDate); err != nil {
	// 	log.Fatalf("Error processing logs: %v", err)
	// }

	r:= router.Router()
	config, err := configs.LoadConfig()
	if err != nil {
		log.Fatalf("Error loading configuration: %v", err)
	}
	port := config.Services.Port
	log.Printf("Starting server on port %d", port)
	if err := http.ListenAndServe(":"+fmt.Sprintf("%d", port), r); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}

}
