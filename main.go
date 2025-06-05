// main.go
package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
	"github.com/phaserunner03/logging/internal/bigquery"
	"github.com/phaserunner03/logging/internal/logs"
)

func init() {
	if err := loadEnvironment(); err != nil {
		log.Fatalf("Failed to initialize environment: %v", err)
	}
}

func loadEnvironment() error {
	if err := godotenv.Load(); err != nil {
		return fmt.Errorf("error loading .env file: %v", err)
	}
	return validateEnvironment()
}

func validateEnvironment() error {
	required := []string{
		"GCP_CREDENTIALS",
		"GCP_PROJECT_ID",
		"BIGQUERY_DATASET_ID",
		"BIGQUERY_TABLE_ID",
	}

	for _, env := range required {
		if os.Getenv(env) == "" {
			return fmt.Errorf("required environment variable %s is not set", env)
		}
	}
	return nil
}

func processLogs(ctx context.Context) error {
	// Fetch logs from Cloud Logging
	entries, err := logs.FetchLogs(ctx)
	if err != nil {
		return fmt.Errorf("failed to fetch logs: %v", err)
	}

	if len(entries) == 0 {
		log.Println("No log entries to process")
		return nil
	}

	// Pre-allocate slice with exact capacity needed
	bqRows := make([]bigquery.BQLogRow, 0, len(entries))
	var conversionErrors int

	// Convert log entries to BigQuery rows in batch
	for _, entry := range entries {
		row, err := logs.ConvertToBQRow(entry)
		if err != nil {
			log.Printf("Warning: Failed to convert log entry: %v", err)
			conversionErrors++
			continue
		}
		bqRows = append(bqRows, row)
	}

	if len(bqRows) == 0 {
		return fmt.Errorf("all %d log entries failed to convert", len(entries))
	}

	// Insert all rows into BigQuery in a single batch
	if err := bigquery.InsertLogs(ctx, bqRows); err != nil {
		return fmt.Errorf("failed to insert logs into BigQuery: %v", err)
	}

	log.Printf("Successfully processed %d log entries (%d conversion errors)", len(bqRows), conversionErrors)
	return nil
}

func main() {
	ctx := context.Background()

	if err := processLogs(ctx); err != nil {
		log.Fatalf("Error processing logs: %v", err)
	}
}
