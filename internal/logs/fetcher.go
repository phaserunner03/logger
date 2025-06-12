package logs

import (
	"context"
	"fmt"
	"log"
	"github.com/phaserunner03/logging/configs"
	logging "cloud.google.com/go/logging/apiv2"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
	"github.com/phaserunner03/logging/internal/analysis"
	"github.com/phaserunner03/logging/internal/bigquery"
	
	logpb "google.golang.org/genproto/googleapis/logging/v2"
)

func ProcessLogs(ctx context.Context, services []string, startDate, endDate string) error {
	// Fetch logs from Cloud Logging
	entries, err := FetchLogs(ctx, services, startDate, endDate)
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
		row, err := ConvertToBQRow(entry)
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


func FetchLogs(ctx context.Context, services []string, startDate, endDate string) ([]*logpb.LogEntry, error) {
	
	config, err := configs.LoadConfig()
	
	if err != nil {
		log.Fatalf("Error loading configuration: %v", err)
	}
	credentials := config.Env.GCP_Credentials
	projectID := config.Env.GCP_ProjectID



	if credentials == "" || projectID == "" {
		return nil, fmt.Errorf("GCP_CREDENTIALS and GCP_PROJECT_ID environment variables must be set")
	}

	logClient, err := logging.NewClient(ctx, option.WithCredentialsFile(credentials))
	if err != nil {
		return nil, fmt.Errorf("failed to create logging client: %v", err)
	}
	defer logClient.Close()

	var entries []*logpb.LogEntry

	for _, service := range services {
		filter := fmt.Sprintf(
			`resource.type="cloud_run_revision" AND resource.labels.service_name="%s" AND timestamp >= "%s" AND timestamp <= "%s"`,
			service, startDate, endDate,
		)

		req := &logpb.ListLogEntriesRequest{
			ResourceNames: []string{"projects/" + projectID},
			Filter:        filter,
			OrderBy:       "timestamp desc",
		}

		it := logClient.ListLogEntries(ctx, req)

		for {
			entry, err := it.Next()
			if err == iterator.Done {
				break
			}
			if err != nil {
				return nil, fmt.Errorf("error iterating log entries: %v", err)
			}
			entries = append(entries, entry)
		}
	}

	return entries, nil
}
