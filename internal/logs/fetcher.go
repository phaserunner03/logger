package logs

import (
	"context"
	"fmt"
	"os"
	"github.com/joho/godotenv"
	
	"cloud.google.com/go/logging/apiv2"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
	logpb "google.golang.org/genproto/googleapis/logging/v2"
)

func FetchLogs(ctx context.Context) ([]*logpb.LogEntry, error) {
	_ = godotenv.Load()
	credentials := os.Getenv("GCP_CREDENTIALS")
	projectID := os.Getenv("GCP_PROJECT_ID")
	filter := `resource.type="cloud_run_revision" AND resource.labels.service_name="loggenerator"`

	if credentials == "" || projectID == "" {
		return nil, fmt.Errorf("GCP_CREDENTIALS and GCP_PROJECT_ID environment variables must be set")
	}

	logClient, err := logging.NewClient(ctx, option.WithCredentialsFile(credentials))
	if err != nil {
		return nil, fmt.Errorf("failed to create logging client: %v", err)
	}
	defer logClient.Close()

	req := &logpb.ListLogEntriesRequest{
		ResourceNames: []string{"projects/" + projectID},
		Filter:        filter,
		OrderBy:       "timestamp desc",
	}

	var entries []*logpb.LogEntry
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

	return entries, nil
}
