package prodsub

import (
	"cloud.google.com/go/pubsub"
	"context"
	"encoding/json"
	"fmt"
	"google.golang.org/api/option"

	"github.com/phaserunner03/logging/configs"
)

type Publisher struct {
	client * pubsub.Client
	topic *pubsub.Topic
	project string
	topicID string
	credPath string
}

func NewPublisher(ctx context.Context, projectID, topicID, credPath string) (* Publisher,error){
	client,err := pubsub.NewClient(ctx,projectID,option.WithCredentialsFile(credPath))
	if err != nil{
		return nil,fmt.Errorf("pubsub client creation failed:%v", err);
	}
	topic:=client.Topic(topicID)
	if topic == nil{
		return nil, fmt.Errorf("pubsub topic not found: %s",topicID)
	}
	return &Publisher{
		client:   client,
		topic:    topic,
		project:  projectID,
		topicID:  topicID,
		credPath: credPath,
	}, nil

}

func (p * Publisher) PublishLog(ctx context.Context, row configs.BQLogRow) error{
	data,err:=json.Marshal(row)
	if err != nil{
		return fmt.Errorf("failed to marshal message:%v",err)
	}
	msg:= &pubsub.Message{
		Data:data,
		Attributes: map[string] string{
			"service_name":row.ServiceName,
		},
	}
	result:= p.topic.Publish(ctx,msg)
	id,err:= result.Get(ctx)

	if err!= nil{
		return fmt.Errorf("failed to publish message:%v", err)
	}
	fmt.Printf("Published message ID: %s\n",id)
	return nil
}

func (p * Publisher) Close(){
	p.topic.Stop()
	p.client.Close()
}
 