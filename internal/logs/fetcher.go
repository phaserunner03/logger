package logs

import (
	"context"
	"fmt"
	"log"
	"github.com/phaserunner03/logging/configs"
	logging "cloud.google.com/go/logging/apiv2"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
	logpb "google.golang.org/genproto/googleapis/logging/v2"
)

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
