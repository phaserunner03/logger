// main.go
package main

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"time"

	"cloud.google.com/go/bigquery"
	"cloud.google.com/go/logging/apiv2"
	"github.com/joho/godotenv"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"

	logpb "google.golang.org/genproto/googleapis/logging/v2"
	logtypepb "google.golang.org/genproto/googleapis/logging/type"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/types/known/structpb"
)

// BQLogRow matches the BigQuery schema (schema.json) provided:
/*
[
  {"name": "timestamp",       "type": "TIMESTAMP", "mode": "REQUIRED"},
  {"name": "severity",        "type": "STRING",    "mode": "NULLABLE"},
  {"name": "log_name",        "type": "STRING",    "mode": "NULLABLE"},
  {"name": "text_payload",    "type": "STRING",    "mode": "NULLABLE"},
  {"name": "json_payload",    "type": "JSON",      "mode": "NULLABLE"},
  {"name": "insert_id",       "type": "STRING",    "mode": "NULLABLE"},
  {"name": "resource_type",   "type": "STRING",    "mode": "NULLABLE"},
  {"name": "resource_labels", "type": "JSON",      "mode": "NULLABLE"},
  {"name": "http_request",    "type": "JSON",      "mode": "NULLABLE"},
  {"name": "trace",           "type": "STRING",    "mode": "NULLABLE"},
  {"name": "span_id",         "type": "STRING",    "mode": "NULLABLE"},
  {"name": "source_location", "type": "JSON",      "mode": "NULLABLE"},
  {"name": "labels",          "type": "JSON",      "mode": "NULLABLE"}
]
*/
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
}

// marshalToJSONString attempts to convert v (various possible types) into a JSON string.
// - If v is nil, returns "null".
// - If v is an empty Struct, HttpRequest, or map[string]string, returns "{}".
// - Otherwise, uses protojson or encoding/json as appropriate.
func marshalToJSONString(v interface{}) string {
	if v == nil {
		return "null"
	}

	var data []byte
	var err error

	switch val := v.(type) {
	case *structpb.Struct:
		if val == nil {
			return "null"
		}
		if len(val.GetFields()) == 0 {
			return "{}"
		}
		marshaller := protojson.MarshalOptions{UseProtoNames: true, EmitUnpopulated: true}
		data, err = marshaller.Marshal(val)
		if err == nil && len(data) == 0 {
			return "{}"
		}

	case *logtypepb.HttpRequest:
		if val == nil {
			return "null"
		}
		marshaller := protojson.MarshalOptions{UseProtoNames: true, EmitUnpopulated: true}
		data, err = marshaller.Marshal(val)
		if err == nil && len(data) == 0 {
			return "{}"
		}

	case *logpb.LogEntrySourceLocation:
		if val == nil {
			return "null"
		}
		marshaller := protojson.MarshalOptions{UseProtoNames: true, EmitUnpopulated: true}
		data, err = marshaller.Marshal(val)
		if err == nil && len(data) == 0 {
			return "{}"
		}

	case map[string]string:
		if val == nil {
			return "null"
		}
		if len(val) == 0 {
			return "{}"
		}
		data, err = json.Marshal(val)

	default:
		data, err = json.Marshal(v)
	}

	if err != nil {
		log.Printf("Error marshalling to JSON: %v. Value: %+v. Returning 'null'.", err, v)
		return "null"
	}
	if len(data) == 0 {
		return "null"
	}
	return string(data)
}

func main() {
	_ = godotenv.Load()

	projectID := os.Getenv("GCP_PROJECT_ID")
	credentialsPath := os.Getenv("GCP_CREDENTIALS")
	datasetID := os.Getenv("BIGQUERY_DATASET_ID")
	tableID := os.Getenv("BIGQUERY_TABLE_ID")

	if projectID == "" || credentialsPath == "" || datasetID == "" || tableID == "" {
		log.Fatalf("Required environment variables missing: GCP_PROJECT_ID, GCP_CREDENTIALS, BIGQUERY_DATASET_ID, BIGQUERY_TABLE_ID")
	}

	ctx := context.Background()

	// Create Logging client
	logClient, err := logging.NewClient(ctx, option.WithCredentialsFile(credentialsPath))
	if err != nil {
		log.Fatalf("Failed to create Logging client: %v", err)
	}
	defer logClient.Close()

	// Create BigQuery client
	bqClient, err := bigquery.NewClient(ctx, projectID, option.WithCredentialsFile(credentialsPath))
	if err != nil {
		log.Fatalf("Failed to create BigQuery client: %v", err)
	}
	defer bqClient.Close()

	inserter := bqClient.Dataset(datasetID).Table(tableID).Inserter()

	// Filter: adjust service_name and timestamp as needed
	filter := `
    resource.type="cloud_run_revision"
    AND resource.labels.service_name="loggenerator"
    AND timestamp >= "2025-01-01T00:00:00Z"
  `
	req := &logpb.ListLogEntriesRequest{
		ResourceNames: []string{"projects/" + projectID},
		Filter:        filter,
		OrderBy:       "timestamp desc",
		PageSize:      100,
	}

	log.Println("Starting to fetch logs and insert into BigQuery…")
	it := logClient.ListLogEntries(ctx, req)

	var processed, inserted int64

	for {
		entry, err := it.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			log.Fatalf("Error iterating log entries: %v", err)
		}
		processed++

		// Initialize row with defaults
		row := BQLogRow{
			Timestamp:      entry.GetTimestamp().AsTime(),
			Severity:       entry.GetSeverity().String(),
			LogName:        entry.GetLogName(),
			TextPayload:    "",
			JsonPayload:    "null",
			InsertID:       entry.GetInsertId(),
			ResourceType:   "",
			ResourceLabels: "null",
			HTTPRequest:    "null",
			Trace:          entry.GetTrace(),
			SpanID:         entry.GetSpanId(),
			SourceLocation: "null",
			Labels:         "null",
		}

		// Resource labels
		if res := entry.GetResource(); res != nil {
			row.ResourceType = res.GetType()
			if labels := res.GetLabels(); labels != nil {
				row.ResourceLabels = marshalToJSONString(labels)
			}
		}

		// HttpRequest
		if hr := entry.GetHttpRequest(); hr != nil {
			row.HTTPRequest = marshalToJSONString(hr)
		}

		// SourceLocation
		if sl := entry.GetSourceLocation(); sl != nil {
			row.SourceLocation = marshalToJSONString(sl)
		}

		// Entry-level labels
		if lbls := entry.GetLabels(); lbls != nil {
			row.Labels = marshalToJSONString(lbls)
		}

		// Payload type
		switch payload := entry.GetPayload().(type) {
		case *logpb.LogEntry_TextPayload:
			row.TextPayload = payload.TextPayload

		case *logpb.LogEntry_JsonPayload:
			row.JsonPayload = marshalToJSONString(payload.JsonPayload)

		case *logpb.LogEntry_ProtoPayload:
			marshaller := protojson.MarshalOptions{
				UseProtoNames:   true,
				EmitUnpopulated: true,
			}
			jsonBytes, mErr := marshaller.Marshal(payload.ProtoPayload)
			if mErr != nil {
				// Build a valid JSON object for the error message
				errMap := map[string]string{"error": mErr.Error()}
				errBytes, _ := json.Marshal(errMap)
				row.JsonPayload = string(errBytes)
			} else {
				js := string(jsonBytes)
				if js == "" || js == "null" {
					row.JsonPayload = "{}"
				} else {
					row.JsonPayload = js
				}
			}

		default:
			// no additional action; TextPayload will remain "" if no text payload
		}

		// Insert into BigQuery
		if err := inserter.Put(ctx, &row); err != nil {
			log.Printf("Failed to insert row into BigQuery: %v. Row: %+v", err, row)
		} else {
			inserted++
		}

		if processed%50 == 0 {
			log.Printf("Processed: %d, Inserted: %d", processed, inserted)
		}
	}

	log.Printf("Finished. Total processed: %d, total inserted: %d", processed, inserted)
}
