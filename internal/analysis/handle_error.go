package analysis

import (
	"context"
	"encoding/json"
	"fmt"
	"google.golang.org/api/option"
	"cloud.google.com/go/pubsub"
	"github.com/phaserunner03/logging/configs"
)

func HandleError(ctx context.Context, bqRows []configs.BQLogRow) error {
	config, err := configs.LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load configuration: %v", err)
	}
	topicID := config.Env.TopicID
	GCP_ProjectID := config.Env.GCP_ProjectID
	credentialsPath := config.Env.GCP_Credentials

	client, err := pubsub.NewClient(ctx, GCP_ProjectID, option.WithCredentialsFile(credentialsPath))
	if err != nil {
		return fmt.Errorf("failed to create Pub/Sub client: %v", err)
	}
	defer client.Close()

	topic := client.Topic(topicID)
	if topic == nil {
		return fmt.Errorf("failed to retrieve Pub/Sub topic: %s", topicID)
	}
	defer topic.Stop()

	for _, row := range bqRows {
		data, err := json.Marshal(row)
		if err != nil {
			return fmt.Errorf("failed to marshal row to JSON: %v", err)
		}
		msg := &pubsub.Message{
			Data: data,
			Attributes: map[string]string{
				"service_name": row.ServiceName,
			},
		}
		result := topic.Publish(ctx, msg)
		// Block until the result is returned and log any errors
		id, err := result.Get(ctx)
		if err != nil {
			return fmt.Errorf("failed to publish message to Pub/Sub: %v", err)
		}
		fmt.Printf("Message published with ID: %s\n", id)
	}

	return nil

}
