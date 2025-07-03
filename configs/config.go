package configs

import (
	"github.com/joho/godotenv"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"os"
	"time"
)

type Config struct {
	Services struct {
		Name     []string `yaml:"name"`
		Severity []string `yaml:"severity"`
	} `yaml:"service"`

	Env struct {
		GCP_Credentials   string
		GCP_ProjectID     string
		BigQueryDatasetID string
		BigQueryTableID   string
		TopicID           string
		SubID             string
		GithubToken       string
		RepoOwner         string
		RepoName          string
		BaseBranch        string
		CodeBasePath      string
		LogFetchType      string
		LogFilePath       string
	}
}

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

func LoadConfig() (*Config, error) {
	filePath := "./configs/services.yaml"
	data, err := ioutil.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	var config Config
	err = yaml.Unmarshal(data, &config)
	if err != nil {
		return nil, err
	}
	err = godotenv.Load()
	if err != nil {
		return nil, err
	}
	config.Env.GCP_Credentials = os.Getenv("GCP_CREDENTIALS")
	config.Env.GCP_ProjectID = os.Getenv("GCP_PROJECT_ID")
	config.Env.BigQueryDatasetID = os.Getenv("BIGQUERY_DATASET_ID")
	config.Env.BigQueryTableID = os.Getenv("BIGQUERY_TABLE_ID")
	config.Env.TopicID = os.Getenv("TOPIC_ID")
	config.Env.SubID = os.Getenv("SUB_ID")
	config.Env.GithubToken = os.Getenv("GITHUB_TOKEN")
	config.Env.RepoOwner = os.Getenv("REPO_OWNER")
	config.Env.RepoName = os.Getenv("REPO_NAME")
	config.Env.BaseBranch = os.Getenv("BASE_BRANCH")
	config.Env.CodeBasePath = os.Getenv("CODE_BASE_PATH")
	config.Env.LogFetchType = os.Getenv("LOG_FETCH_TYPE")
	config.Env.LogFilePath = os.Getenv("LOG_FILE_PATH")

	return &config, nil
}
