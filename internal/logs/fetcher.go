package logs

import (
	"context"
	"fmt"
	"os"
	"strings"
	"bufio"
	"regexp"
	"log"
	"time"
	"encoding/json"
	"github.com/phaserunner03/logging/configs"
	logging "cloud.google.com/go/logging/apiv2"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
	logpb "google.golang.org/genproto/googleapis/logging/v2"
)

func FetchLogsFromCloud(ctx context.Context, services []string, startDate, endDate string) ([]*logpb.LogEntry, error) {
	
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

func FetchLogsFromFile(ctx context.Context, filePath string) ([]*configs.BQLogRow, error){
	file,err := os.Open(filePath)
	if os.IsNotExist(err) {
		return nil, fmt.Errorf("log file does not exist: %s", filePath)
	}
	if err!=nil{
		return nil, fmt.Errorf("failed to open log file: %v", err)
	}
	defer file.Close()

	var entries []*configs.BQLogRow
	var currentLog strings.Builder

	scanner := bufio.NewScanner(file)
	logStartPattern := regexp.MustCompile(`^\{timestamp:`)

	for scanner.Scan() {
		line := scanner.Text()
		
		if logStartPattern.MatchString(line) {
			// Flush previous entry
			if currentLog.Len() > 0 {
				entry, err := parseLogEntry(currentLog.String())
				if err == nil && entry != nil {
					entries = append(entries, entry)
				}
			}
			currentLog.Reset()
		}
		currentLog.WriteString(line + "\n")
	}

	if currentLog.Len() > 0 {
		entry, err := parseLogEntry(currentLog.String())
		if err == nil && entry != nil {
			entries = append(entries, entry)
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error while scanning file: %v", err)
	}

	return entries, nil
}



func parseLogEntry(raw string) (*configs.BQLogRow, error) {
	// Remove outer braces if present
	raw = strings.TrimSpace(raw)
	if strings.HasPrefix(raw, "{") && strings.HasSuffix(raw, "}") {
		raw = raw[1:len(raw)-1]
	}

	// Split by commas outside of braces (naive but effective for simple logs)
	parts := strings.Split(raw, ",")
	parsed := make(map[string]interface{})

	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		colonIndex := strings.Index(part, ":")
		if colonIndex == -1 {
			continue
		}

		key := strings.TrimSpace(part[:colonIndex])
		value := strings.TrimSpace(part[colonIndex+1:])

		// Clean up key and value
		key = strings.Trim(key, `"`)
		if !strings.HasPrefix(key, `"`) {
			key = `"` + key + `"`
		}

		// Handle different value types
		var finalValue interface{}
		if strings.HasPrefix(value, "[") || strings.HasPrefix(value, "{") {
			finalValue = value
		} else {
			value = strings.Trim(value, `"`)
			finalValue = value
		}

		// Use json.Unmarshal for array or object
		if s, ok := finalValue.(string); ok && (strings.HasPrefix(s, "[") || strings.HasPrefix(s, "{")) {
			var v interface{}
			err := json.Unmarshal([]byte(s), &v)
			if err == nil {
				finalValue = v
			}
		}

		parsed[key[1:len(key)-1]] = finalValue // remove quotes from key
	}

	// Optional: Convert back to JSON string for JsonPayload
	jsonLike, err := json.Marshal(parsed)
	if err != nil {
		return nil, fmt.Errorf("failed to re-marshal log: %v", err)
	}

	// Extract values
	return &configs.BQLogRow{
		Timestamp:   parseTimestamp(parsed["timestamp"]),
		Severity:    fmt.Sprintf("%v", parsed["level"]),
		JsonPayload: string(jsonLike),
		ServiceName: fmt.Sprintf("%v", parsed["function"]),
	}, nil
}




func parseTimestamp(ts interface{}) time.Time {
	str, ok := ts.(string)
	if !ok {
		return time.Now()
	}
	t, err := time.Parse("2006-01-02 15:04:05", str)
	if err != nil {
		return time.Now()
	}
	return t
}