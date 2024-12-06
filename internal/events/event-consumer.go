package events

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sqs"
	"go-test/internal/handlers"
	"go-test/internal/model"
	"go.temporal.io/sdk/client"
	"go.uber.org/zap"
	"log"
	"sync"
)

type EventConsumerConfig struct {
	Logger         *zap.Logger
	Endpoint       string
	TemporalClient client.Client
	TaskQueueName  string
}

func NewEventConsumerConfig(logger *zap.Logger, temporalClient client.Client, taskQueueName string) *EventConsumerConfig {
	return &EventConsumerConfig{
		Logger:         logger,
		Endpoint:       "http://localhost:4566",
		TemporalClient: temporalClient,
		TaskQueueName:  taskQueueName,
	}
}

type EventConsumer struct {
	sqsSvc   *sqs.SQS
	queueURL string
	logger   *zap.Logger
	ctx      context.Context
	cancel   context.CancelFunc
	wg       sync.WaitGroup
	Handler  *handlers.DeliveryEventConsumer
}

func (c *EventConsumerConfig) InitEventConsumer(queueName string) *EventConsumer {
	ctx, cancel := context.WithCancel(context.Background())

	sess, err := session.NewSession(&aws.Config{
		Region:   aws.String("us-east-1"),
		Endpoint: aws.String(c.Endpoint),
	})
	if err != nil {
		log.Fatalf("failed to create session: %v", err)
	}

	sqsSvc := sqs.New(sess)

	return &EventConsumer{
		sqsSvc:   sqsSvc,
		queueURL: fmt.Sprintf("%s/000000000000/%s", c.Endpoint, queueName),
		logger:   c.Logger,
		ctx:      ctx,
		cancel:   cancel,
		Handler:  handlers.NewDeliveryEventConsumer(c.Logger, c.TemporalClient, c.TaskQueueName),
	}
}

func (ec *EventConsumer) StartConsuming() {
	ec.wg.Add(1)
	defer ec.wg.Done()

	for {
		select {
		case <-ec.ctx.Done():
			ec.logger.Info("Shutting down consumer...")
			return
		default:
			result, err := ec.sqsSvc.ReceiveMessage(&sqs.ReceiveMessageInput{
				QueueUrl:            aws.String(ec.queueURL),
				MaxNumberOfMessages: aws.Int64(1),
				WaitTimeSeconds:     aws.Int64(10),
			})
			if err != nil {
				continue
			}

			if len(result.Messages) == 0 {
				continue
			}

			for _, message := range result.Messages {
				ec.logger.Info("Received message", zap.String("message", *message.Body))

				var event model.DeliveryPackage
				err := json.Unmarshal([]byte(*message.Body), &event)
				if err != nil {
					ec.logger.Error("Failed to unmarshal message", zap.Error(err))
					return
				}

				err = ec.Handler.Handle(ec.ctx, &event)
				if err != nil {
					ec.logger.Error("Failed to handle message", zap.Error(err))
					return
				}

				_, err = ec.sqsSvc.DeleteMessage(&sqs.DeleteMessageInput{
					QueueUrl:      aws.String(ec.queueURL),
					ReceiptHandle: message.ReceiptHandle,
				})
				if err != nil {
					ec.logger.Error("failed to delete message", zap.Error(err))
				} else {
					ec.logger.Info("Message deleted", zap.String("message", *message.Body))
				}
			}
		}
	}
}

func (ec *EventConsumer) Dispose() {
	ec.cancel()
	ec.wg.Wait()
	ec.logger.Info("Consumer has been disposed")
}
