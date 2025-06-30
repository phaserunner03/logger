package analysis

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"

	"github.com/phaserunner03/logging/configs"
	"github.com/phaserunner03/logging/internal/prodsub"
)

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
		return "", fmt.Errorf("flask returned status %d: %s", resp.StatusCode, string(body))
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

	for _, row := range errorLogs {
		if err := publisher.PublishLog(ctx, row); err != nil {
			return fmt.Errorf("failed to publish message: %v", err)
		}
	}

	logQueue := make(chan configs.BQLogRow, 10)
	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()
		for row := range logQueue {
			fmt.Printf("Received log from service %s with severity %s: %s\n",
				row.ServiceName, row.Severity, row.TextPayload)

			// suggestion, err := SuggestFix(row.Timestamp.Format(time.RFC3339), row.TextPayload)
			// if err != nil {
				// fmt.Printf("failed to get suggestion: %v\n", err)
				// continue
			// }

			// fmt.Printf("✅ Suggested Fix:\n%s\n", suggestion)
		}
	}()

	// Create subscriber
	subscriber, err := prodsub.NewSubscriber(ctx, GCP_ProjectID, subID, credentialsPath)
	if err != nil {
		return fmt.Errorf("failed to create Pub/Sub subscriber: %v", err)
	}
	defer subscriber.Close()

	// Use timeout to stop listening after a short idle period
	timeoutCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	err = subscriber.Listen(timeoutCtx, func(row configs.BQLogRow) error {
		logQueue <- row
		return nil
	})

	close(logQueue)
	wg.Wait()

	return err
}

func parseErrorLogs(bqRows []configs.BQLogRow) ([]configs.BQLogRow, error) {
	var errorLogs []configs.BQLogRow
	for _, row := range bqRows {
		fmt.Print(row)
		if row.Severity == "ERROR" {
			errorLogs = append(errorLogs, row)
		}
	}
	return errorLogs, nil
}
