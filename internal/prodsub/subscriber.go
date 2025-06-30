package prodsub

import (
	"context"
	"encoding/json"
	"fmt"
	"time"
	"cloud.google.com/go/pubsub"
	"github.com/phaserunner03/logging/configs"
	"google.golang.org/api/option"
)

type Subscriber struct {
	client *pubsub.Client
	subscriptionID string
	subcription *pubsub.Subscription
	credPath string
}

func NewSubscriber(ctx context.Context, projectID, subscriptionID, credPath string) (*Subscriber, error) {
	client, err := pubsub.NewClient(ctx, projectID, option.WithCredentialsFile(credPath))
	if err != nil {
		return nil, fmt.Errorf("pubsub subscription not found: %v", err)
	}

	sub := client.Subscription(subscriptionID)

	// Prevent endless streaming; make it timeout-able
	sub.ReceiveSettings.MaxOutstandingMessages = 10
	sub.ReceiveSettings.MaxExtension = 10 * time.Second

	return &Subscriber{
		client:         client,
		subscriptionID: subscriptionID,
		subcription:    sub,
		credPath:       credPath,
	}, nil
}


func (s *Subscriber) Listen(ctx context.Context, handler func(configs.BQLogRow) error) error {
	
	fmt.Println("Subscriber started listening...")
	return s.subcription.Receive(ctx, func(ctx context.Context, msg * pubsub.Message){
		var row configs.BQLogRow
		err:=json.Unmarshal(msg.Data,&row)
		if err!=nil{
			fmt.Printf("Failed to unmarshal the message:%v\n",err)
			msg.Nack()
			return
		}
		//process the message calling ai/ml service
		err = handler(row)
		if err!=nil{
			fmt.Printf("handler error:%v\n",err)
			msg.Nack()
			return
		}
		msg.Ack()
		fmt.Printf("message processed and acked. Service;%s\n", row.ServiceName)
	})
}

func (s *Subscriber) Close(){
	_ = s.client.Close()
}
