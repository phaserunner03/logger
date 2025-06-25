package analysis

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/phaserunner03/logging/configs"
	"github.com/phaserunner03/logging/internal/prodsub"
)

// SuggestFix calls the Flask service to get suggested fix
func SuggestFix(timestamp, errorMessage string) (string, error) {
	payload := map[string]string{
		"timestamp":     timestamp,
		"error_message": errorMessage,
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("failed to marshal JSON: %v", err)
	}

	resp, err := http.Post("http://127.0.0.1:8080/suggest-fix", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("POST request failed: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response body: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("Flask returned status %d: %s", resp.StatusCode, string(body))
	}

	var parsed map[string]interface{}
	err = json.Unmarshal(body, &parsed)
	if err != nil {
		return "", fmt.Errorf("JSON unmarshal failed: %v", err)
	}

	fix, ok := parsed["suggested_fix"].(string)
	if !ok {
		return "", fmt.Errorf("suggested_fix not found or invalid in response: %v", parsed)
	}

	return fix, nil
}

func HandleError(ctx context.Context, bqRows []configs.BQLogRow) error {
	config, err := configs.LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load configuration: %v", err)
	}

	topicID := config.Env.TopicID
	subID := config.Env.SubID
	GCP_ProjectID := config.Env.GCP_ProjectID
	credentialsPath := config.Env.GCP_Credentials

	errorLogs, err := parseErrorLogs(bqRows)
	if err != nil {
		return fmt.Errorf("failed to parse error logs: %v", err)
	}
	if len(errorLogs) == 0 {
		fmt.Println("No error logs found, skipping Pub/Sub publishing")
		return nil
	}

	publisher, err := prodsub.NewPublisher(ctx, GCP_ProjectID, topicID, credentialsPath)
	if err != nil {
		return fmt.Errorf("failed to create Pub/Sub client: %v", err)
	}
	defer publisher.Close()

	for _, row := range bqRows {
		if err := publisher.PublishLog(ctx, row); err != nil {
			return fmt.Errorf("failed to publish message: %v", err)
		}
	}

	// Create a log queue with buffer
	logQueue := make(chan configs.BQLogRow, 10)

	// Start a single worker to process logs serially
	go func() {
		for row := range logQueue {
			fmt.Printf("Received log from service %s with severity %s: %s\n",
				row.ServiceName, row.Severity, row.TextPayload)

			suggestion, err := SuggestFix(row.Timestamp.Format(time.RFC3339), row.TextPayload)
			if err != nil {
				fmt.Printf("failed to get suggestion: %v\n", err)
				continue
			}

			fmt.Printf("✅ Suggested Fix:\n%s\n", suggestion)
		}
	}()

	subscriber, err := prodsub.NewSubscriber(ctx, GCP_ProjectID, subID, credentialsPath)
	if err != nil {
		return fmt.Errorf("failed to create Pub/Sub subscriber: %v", err)
	}
	defer subscriber.Close()

	// Enqueue logs into the logQueue for serial processing
	err = subscriber.Listen(ctx, func(row configs.BQLogRow) error {
		logQueue <- row
		return nil
	})

	return err
}

func parseErrorLogs(bqRows []configs.BQLogRow) ([]configs.BQLogRow, error) {
	var errorLogs []configs.BQLogRow
	for _, row := range bqRows {
		if row.Severity == "ERROR" {
			errorLogs = append(errorLogs, row)
		}
	}
	return errorLogs, nil
}
