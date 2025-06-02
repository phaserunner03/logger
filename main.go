package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"cloud.google.com/go/logging/apiv2"
	"github.com/joho/godotenv"
	logpb "google.golang.org/genproto/googleapis/logging/v2"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
)

func main() {
	// Load .env file
	if err := godotenv.Load(); err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}

	projectID := os.Getenv("GCP_PROJECT_ID")
	credentials := os.Getenv("GCP_CREDENTIALS")

	if projectID == "" || credentials == "" {
		log.Fatalf("Missing environment variables")
	}

	ctx := context.Background()

	client, err := logging.NewClient(ctx, option.WithCredentialsFile(credentials))
	if err != nil {
		log.Fatalf("Failed to create logging client: %v", err)
	}
	defer client.Close()

	// Open file for writing logs
	file, err := os.Create("cloud_run_logs.txt")
	if err != nil {
		log.Fatalf("Failed to create log file: %v", err)
	}
	defer file.Close()

	// Filter for Cloud Run logs
	filter := `
		resource.type="cloud_run_revision"
		resource.labels.service_name="backend"
	`

	req := &logpb.ListLogEntriesRequest{
		ResourceNames: []string{"projects/" + projectID},
		Filter:        filter,
		OrderBy:       "timestamp desc",
	}

	it := client.ListLogEntries(ctx, req)

	for {
		entry, err := it.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			log.Fatalf("Error iterating logs: %v", err)
		}

		ts := entry.GetTimestamp().AsTime().Format(time.RFC3339)
		var line string

		switch payload := entry.Payload.(type) {
		case *logpb.LogEntry_TextPayload:
			line = fmt.Sprintf("[%s] %s\n", ts, payload.TextPayload)
		case *logpb.LogEntry_JsonPayload:
			line = fmt.Sprintf("[%s] JSON Payload: %v\n", ts, payload.JsonPayload)
		default:
			line = fmt.Sprintf("[%s] (non-text payload)\n", ts)
		}

		_, err = file.WriteString(line)
		if err != nil {
			log.Printf("Failed to write log entry: %v", err)
		}
	}

	fmt.Println("Log collection complete.")
}
