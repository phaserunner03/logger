package bigquery

import (
	"context"
	"fmt"
	"log"
	"github.com/phaserunner03/logging/configs"
	"cloud.google.com/go/bigquery"
	"google.golang.org/api/option"
)


// InsertLogs inserts log entries into BigQuery
func InsertLogs(ctx context.Context, rows []configs.BQLogRow) error {
	if len(rows) == 0 {
		return nil
	}

	
	config, err := configs.LoadConfig()
	if err != nil {
		log.Fatalf("Error loading configuration: %v", err)
	}
	projectID := config.Env.GCP_ProjectID
	credentialsPath := config.Env.GCP_Credentials
	datasetID := config.Env.BigQueryDatasetID
	tableID := config.Env.BigQueryTableID

	if projectID == "" || credentialsPath == "" || datasetID == "" || tableID == "" {
		return fmt.Errorf("required environment variables are not set")
	}

	client, err := bigquery.NewClient(ctx, projectID, option.WithCredentialsFile(credentialsPath))
	if err != nil {
		return fmt.Errorf("failed to create BigQuery client: %v", err)
	}
	defer client.Close()

	inserter := client.Dataset(datasetID).Table(tableID).Inserter()
	if err := inserter.Put(ctx, rows); err != nil {
		return fmt.Errorf("failed to insert rows: %v", err)
	}

	return nil
}
