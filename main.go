// main.go
package main

import (
	"context"
	"fmt"
	"log"

	"github.com/phaserunner03/logging/configs"
	"github.com/phaserunner03/logging/internal/analysis"
	"github.com/phaserunner03/logging/internal/bigquery"

	"github.com/phaserunner03/logging/internal/logs"

	"github.com/robfig/cron/v3"
	"time"
)

func processLogsFromFile(ctx context.Context, filePath string) error {

	entries, err := logs.FetchLogsFromFile(ctx, filePath)
	if err != nil {
		return fmt.Errorf("failed to fetch logs from file: %v", err)
	}

	if len(entries) == 0 {
		log.Println("No log entries to process")
		return nil
	}
		
	log.Println("📋 Showing sample logs (max 10):")
	for i := 0; i < len(entries) && i < 10; i++ {
		entry := entries[i]
		log.Printf("[%d] Timestamp: %s | Function: %s | Message: %s\n",
			i+1, entry.Timestamp.Format("2006-01-02 15:04:05"), entry.ServiceName, entry.JsonPayload)
	}


	fmt.Printf("successfully processed log entries from file")
	return nil
}

func processLogsFromCloud(ctx context.Context, services []string, startDate, endDate string) error {
	// Fetch logs from Cloud Logging
	entries, err := logs.FetchLogsFromCloud(ctx, services, startDate, endDate)
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
	ctx := context.Background()
	config, err := configs.LoadConfig()
	if err != nil {
		log.Fatalf("Error loading configuration: %v", err)
	}
	services := config.Services.Name

	switch config.Env.LogFetchType {
	case "file":
		fmt.Printf("Fetching logs from file: %v\n", config.Services.Name)
		err := processLogsFromFile(ctx, config.Env.LogFilePath)
		if err != nil {
			log.Fatalf("Error processing logs from file: %v", err)
		}

	case "cloud":
		fmt.Printf("Fetching logs from cloud %v\n", config.Services.Name)
		c := cron.New(cron.WithLocation(time.UTC))

		c.AddFunc("* * * * *", func() {
			end := time.Now().UTC()
			start := end.Add(-1 * time.Minute)
			startDate := start.Format(time.RFC3339)
			endDate := end.Format(time.RFC3339)

			log.Printf("Starting log processing from %s to %s", startDate, endDate)
			if err := processLogsFromCloud(ctx, services, startDate, endDate); err != nil {
				log.Printf("Error: %v", err)
			}
		})
		c.Start()
		select {}

	}

}
