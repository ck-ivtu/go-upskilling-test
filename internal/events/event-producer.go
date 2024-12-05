package events

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sqs"
	"go.uber.org/zap"
	"log"
	"os"
	"strings"
)

type EventProducerConfig struct {
	Logger   *zap.Logger
	Endpoint string
}

func NewEventProducerConfig(logger *zap.Logger) *EventProducerConfig {
	return &EventProducerConfig{
		Logger:   logger,
		Endpoint: "http://localhost:4566",
	}
}

type EventProducer struct {
	sqsSvc   *sqs.SQS
	queueURL string
	logger   *zap.Logger
	Endpoint string
}

func (c *EventProducerConfig) InitEventProducer(queueName string) *EventProducer {
	// TODO: Add config
	os.Setenv("AWS_ACCESS_KEY_ID", "test")
	os.Setenv("AWS_SECRET_ACCESS_KEY", "test")
	os.Setenv("AWS_DEFAULT_REGION", "us-east-1")

	sess, err := session.NewSession(&aws.Config{
		Region:   aws.String("us-east-1"),
		Endpoint: aws.String(c.Endpoint),
	})
	if err != nil {
		log.Fatalf("failed to create session: %v", err)
	}

	sqsSvc := sqs.New(sess)

	producer := &EventProducer{
		sqsSvc:   sqsSvc,
		queueURL: fmt.Sprintf("%s/000000000000/%s", c.Endpoint, queueName),
		logger:   c.Logger,
		Endpoint: c.Endpoint,
	}

	producer.CreateQueueIfNotExists(queueName)

	return producer
}

func (ep *EventProducer) CreateQueueIfNotExists(queueName string) {
	result, err := ep.sqsSvc.ListQueues(&sqs.ListQueuesInput{})
	if err != nil {
		ep.logger.Error("failed to list queues", zap.Error(err))
		return
	}

	queueExists := false
	for _, url := range result.QueueUrls {
		if strings.HasSuffix(*url, queueName) {
			queueExists = true
			ep.queueURL = *url
			break
		}
	}

	if !queueExists {
		ep.logger.Info("Queue does not exist, creating queue", zap.String("queueName", queueName))
		_, err := ep.sqsSvc.CreateQueue(&sqs.CreateQueueInput{
			QueueName: aws.String(queueName),
		})
		if err != nil {
			ep.logger.Error("failed to create queue", zap.Error(err))
			return
		}

		ep.queueURL = fmt.Sprintf("%s/000000000000/%s", ep.Endpoint, queueName)
		ep.logger.Info("Queue created successfully", zap.String("queueName", queueName))
	}
}

func (ep *EventProducer) SendEvent(message string) error {
	_, err := ep.sqsSvc.SendMessage(&sqs.SendMessageInput{
		QueueUrl:    aws.String(ep.queueURL),
		MessageBody: aws.String(message),
	})
	if err != nil {
		ep.logger.Error("failed to send message", zap.Error(err))
		return err
	}

	ep.logger.Info("Message sent successfully", zap.String("message", message))

	return nil
}
