package configs

import (
	"io/ioutil"
	"os"
	"github.com/joho/godotenv"
	"gopkg.in/yaml.v2"
)

type Config struct {
	Services struct {
		Name []string `yaml:"name"`
	} `yaml:"service"`

	Env struct {
		GCP_Credentials   string
		GCP_ProjectID     string
		BigQueryDatasetID string
		BigQueryTableID   string
	}
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

	return &config, nil
}
