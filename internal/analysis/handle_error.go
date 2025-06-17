package analysis


import (
	"context"
	"fmt"
	"github.com/phaserunner03/logging/configs"
	"github.com/phaserunner03/logging/internal/prodsub"
)

func HandleError(ctx context.Context, bqRows []configs.BQLogRow) error {
	config, err := configs.LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load configuration: %v", err)
	}
	topicID := config.Env.TopicID
	GCP_ProjectID := config.Env.GCP_ProjectID
	credentialsPath := config.Env.GCP_Credentials

	errorLogs, err:= parseErrorLogs(bqRows)
	if err != nil {
		return fmt.Errorf("failed to parse error logs: %v", err)
	}
	if len(errorLogs) == 0 {
		fmt.Println("No error logs found, skipping Pub/Sub publishing")
		return nil
	}

	

	publisher,err := prodsub.NewPublisher(ctx,GCP_ProjectID,topicID,credentialsPath)
	if err != nil {
		return fmt.Errorf("failed to create Pub/Sub client: %v", err)
	}
	
	defer publisher.Close()

	
	for _, row := range bqRows {
		if err:= publisher.PublishLog(ctx,row); err!= nil{
			return fmt.Errorf("failed to publish message: %v",err)
		}
	}

	return nil

}


func parseErrorLogs(bqRows []configs.BQLogRow) ([]configs.BQLogRow, error) {
	var errorLogs []configs.BQLogRow
	for _,row:= range bqRows{
		if row.Severity =="ERROR"{
			errorLogs = append(errorLogs, row)
		}
	}
	
	return errorLogs, nil
}