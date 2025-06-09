package bigquery

import (
	"context"
	"fmt"
	"os"
	"time"

	"cloud.google.com/go/bigquery"
	"google.golang.org/api/option"
)

type BQLogRow struct {
	Timestamp      time.Time `bigquery:"timestamp"`       // REQUIRED
	Severity       string    `bigquery:"severity"`        // NULLABLE
	LogName        string    `bigquery:"log_name"`        // NULLABLE
	TextPayload    string    `bigquery:"text_payload"`    // NULLABLE
	JsonPayload    string    `bigquery:"json_payload"`    // NULLABLE (JSON type in BigQuery)
	InsertID       string    `bigquery:"insert_id"`       // NULLABLE
	ResourceType   string    `bigquery:"resource_type"`   // NULLABLE
	ResourceLabels string    `bigquery:"resource_labels"` // NULLABLE (JSON type)
	HTTPRequest    string    `bigquery:"http_request"`    // NULLABLE (JSON type)
	Trace          string    `bigquery:"trace"`           // NULLABLE
	SpanID         string    `bigquery:"span_id"`         // NULLABLE
	SourceLocation string    `bigquery:"source_location"` // NULLABLE (JSON type)
	Labels         string    `bigquery:"labels"`          // NULLABLE (JSON type)
	ServiceName    string    `bigquery:"service_name"`    // NULLABLE // Added service name field
}

// InsertLogs inserts log entries into BigQuery
func InsertLogs(ctx context.Context, rows []BQLogRow) error {
	if len(rows) == 0 {
		return nil
	}

	projectID := os.Getenv("GCP_PROJECT_ID")
	credentialsPath := os.Getenv("GCP_CREDENTIALS")
	datasetID := os.Getenv("BIGQUERY_DATASET_ID")
	tableID := os.Getenv("BIGQUERY_TABLE_ID")

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
