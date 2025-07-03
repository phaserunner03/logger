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
	"path/filepath"
	"os"
	"github.com/phaserunner03/logging/configs"
	githubconnector "github.com/phaserunner03/logging/internal/github"
	"github.com/phaserunner03/logging/internal/prodsub"
)

type FixResponse struct {
	Filename string  `json:"filename"`
	Changes string `json:"changes"`
	Explanation string `json:"explanation"`
	RawText string `json:"raw_text"`
}

type SuggestFixRequest struct {
	Timestamp    string `json:"timestamp"`
	ErrorMessage string `json:"error_message"`
}

func SuggestFix(ctx context.Context, timestamp, rawErrorMessage string) (*FixResponse, error) {
	// Build structured and strongly-typed payload
	var outer struct {
		Message string `json:"message"`
	}
	if err := json.Unmarshal([]byte(rawErrorMessage), &outer); err != nil {
		return nil, fmt.Errorf("failed to unmarshal outer error message: %v", err)
	}

	// Step 2: Unmarshal inner JSON inside the "message" string
	var inner struct {
		Message string `json:"message"`
	}
	if err := json.Unmarshal([]byte(outer.Message), &inner); err != nil {
		return nil, fmt.Errorf("failed to unmarshal inner error message: %v", err)
	}
	payload := SuggestFixRequest{
		Timestamp:    timestamp,
		ErrorMessage: inner.Message,
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal JSON: %v", err)
	}

	// Prepare HTTP request with context
	req, err := http.NewRequestWithContext(ctx, "POST", "http://127.0.0.1:8080/suggest-fix", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")

	// Send request
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("POST request failed: %v", err)
	}
	defer resp.Body.Close()

	// Read body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %v", err)
	}

	// Handle non-200 response
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("flask returned status %d: %s", resp.StatusCode, string(body))
	}

	// Parse response
	var fixResponse FixResponse
	if err := json.Unmarshal(body, &fixResponse); err != nil {
		return nil, fmt.Errorf("JSON unmarshal failed: %v", err)
	}

	// Validate content
	if fixResponse.Filename == "" || len(fixResponse.Changes) == 0 {
		return nil, fmt.Errorf("suggested fix not found or invalid in response: %v", fixResponse)
	}

	return &fixResponse, nil
}


// func SuggestFix(ctx context.Context,timestamp, errorMessage string) (* FixResponse, error){
// 	payload := map[string]string{
// 		"timestamp":     timestamp,
// 		"error_message": errorMessage,
// 	}
// 	jsonData,err:= json.Marshal(payload)
// 	if err != nil {
// 		return nil, fmt.Errorf("failed to marshal JSON: %v", err)
// 	}

// 	req, err := http.NewRequestWithContext(ctx, "POST", "http://127.0.0.1:8080/suggest-fix", bytes.NewBuffer(jsonData))
// 	if err != nil {
// 		return nil, fmt.Errorf("failed to create HTTP request: %v", err)
// 	}
// 	req.Header.Set("Content-Type", "application/json")

// 	resp,err := http.DefaultClient.Do(req)
// 	if err != nil {
// 		return nil, fmt.Errorf("POST request failed: %v", err)
// 	}
// 	defer resp.Body.Close()

// 	body,err := io.ReadAll(resp.Body)
// 	if err != nil {
// 		return nil, fmt.Errorf("failed to read response body: %v", err)
// 	}
	
// 	if resp.StatusCode != http.StatusOK {
// 		return nil, fmt.Errorf("flask returned status %d: %s", resp.StatusCode, string(body))
// 	}
// 	var fixResponse FixResponse
// 	err = json.Unmarshal(body, &fixResponse)
// 	if err != nil {
// 		return nil, fmt.Errorf("JSON unmarshal failed: %v", err)
// 	}
	
// 	if fixResponse.Filename == "" || len(fixResponse.Changes) == 0 {
// 		return nil, fmt.Errorf("suggested fix not found or invalid in response: %v", fixResponse)
// 	}
// 	return &fixResponse, nil
// }

func HandleError(ctx context.Context, bqRows []configs.BQLogRow) error {
	cfg,err:= configs.LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load configuration: %v", err)
	}

	errorLogs,err:= parseErrorLogs(bqRows)
	if err != nil {
		return fmt.Errorf("failed to parse error logs: %v", err)
	}
	if len(errorLogs) == 0 {
		fmt.Println("No error logs found, skipping Pub/Sub publishing")
		return nil
	}
	
	if err:= publishLogs(ctx,cfg,errorLogs); err!=nil{
		return fmt.Errorf("failed to publish logs: %v", err)
	}
	return processLogs(ctx, cfg)

}

func writeFullFile(filePath, content string) error {
	return os.WriteFile(filePath, []byte(content), 0644)
}

func sanitizeBranchName(name string) string {
	return filepath.Base(name) 
}

func buildPRConfig(cfg *configs.Config, fix *FixResponse, filePath string) githubconnector.PRConfig {
	return githubconnector.PRConfig{
		GithubToken:   cfg.Env.GithubToken,
		RepoOwner:     cfg.Env.RepoOwner,
		RepoName:      cfg.Env.RepoName,
		BaseBranch:    cfg.Env.BaseBranch,
		NewBranch:     fmt.Sprintf("fix-%s-%d", sanitizeBranchName(fix.Filename), time.Now().Unix()),
		FixFilePath:   filePath,
		LocalRepoPath: cfg.Env.CodeBasePath,
		CommitMessage: fmt.Sprintf("Fix: %s", fix.Filename),
		PRTitle:       fmt.Sprintf("Fix for %s", fix.Filename),
		PRBody:        fix.Explanation,
	}
}

func applyFix(ctx context.Context, cfg *configs.Config, fix *FixResponse) error {
	if fix == nil {
		return fmt.Errorf("no fix provided")
	}
	
	fixFilePath := filepath.Join(cfg.Env.CodeBasePath, fix.Filename)
	fmt.Printf("📝 Writing fix to file: %s\n", fixFilePath)

	if err := writeFullFile(fixFilePath, fix.Changes); err != nil {
		return fmt.Errorf("failed to write fix: %v", err)
	}

	// Create a PR
	prCfg := buildPRConfig(cfg, fix, fixFilePath)
	if err := githubconnector.CreatePR(prCfg); err != nil {
		return fmt.Errorf("failed to create PR: %v", err)
	}

	fmt.Println("✅ PR successfully created")
	return nil
}


func publishLogs(ctx context.Context, cfg *configs.Config, logs []configs.BQLogRow) error {
	publisher,err := prodsub.NewPublisher(ctx, cfg.Env.GCP_ProjectID, cfg.Env.TopicID, cfg.Env.GCP_Credentials)
	if err != nil {
		return fmt.Errorf("failed to create Pub/Sub client: %v", err)
	}
	defer publisher.Close()

	for _, row := range logs {
		if err := publisher.PublishLog(ctx, row); err != nil {
			return fmt.Errorf("failed to publish message: %v", err)
		}
	}

	return nil
}


func processLogs(ctx context.Context, cfg *configs.Config) error {
	logQueue := make(chan configs.BQLogRow, 10)
	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()
		for row := range logQueue {
			fmt.Printf("🔍 Log from %s [%s]: %s\n", row.ServiceName, row.Severity, row.JsonPayload)

			fix, err := SuggestFix(ctx, row.Timestamp.Format(time.RFC3339), row.JsonPayload)
			if err != nil {
				fmt.Printf("Failed to suggest fix: %v\n", err)
				continue
			}

			if err := applyFix(ctx, cfg, fix); err != nil {
				fmt.Printf("Failed to apply fix & PR: %v\n", err)
			} else {
				fmt.Println("✅ Pull Request successfully created.")
			}
		}
	}()

	subscriber, err := prodsub.NewSubscriber(ctx, cfg.Env.GCP_ProjectID, cfg.Env.SubID, cfg.Env.GCP_Credentials)
	if err != nil {
		return fmt.Errorf("failed to create subscriber: %v", err)
	}
	defer subscriber.Close()

	timeoutCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	if err := subscriber.Listen(timeoutCtx, func(row configs.BQLogRow) error {
		logQueue <- row
		return nil
	}); err != nil {
		return fmt.Errorf("failed to listen: %v", err)
	}

	close(logQueue)
	wg.Wait()
	return nil
}
	


func parseErrorLogs(bqRows []configs.BQLogRow) ([]configs.BQLogRow, error) {
    cfg, _ := configs.LoadConfig()
    severities := cfg.Services.Severity

    severitySet := make(map[string]struct{}, len(severities))
    for _, sev := range severities {
        severitySet[sev] = struct{}{}
    }

    var filteredLogs []configs.BQLogRow
    for _, row := range bqRows {
        if _, exists := severitySet[row.Severity]; exists {
            filteredLogs = append(filteredLogs, row)
        }
    }

    return filteredLogs, nil
}