package handlers

import (
	"context"
	"go-test/internal/model"
	"go-test/internal/workflow"
	"go.temporal.io/sdk/client"
	"go.uber.org/zap"
)

const (
	PackageDeliveryQueueName = "package-delivery-queue"
)

type DeliveryEventConsumer struct {
	eventName                    string
	Logger                       *zap.Logger
	TemporalClient               client.Client
	PackageDeliveryTaskQueueName string
}

func NewDeliveryEventConsumer(logger *zap.Logger, temporalClient client.Client, packageDeliveryTaskQueueName string) *DeliveryEventConsumer {
	return &DeliveryEventConsumer{
		Logger:                       logger,
		TemporalClient:               temporalClient,
		PackageDeliveryTaskQueueName: packageDeliveryTaskQueueName,
	}
}

func (d *DeliveryEventConsumer) Handle(ctx context.Context, deliveryPackage *model.DeliveryPackage) error {
	wo := client.StartWorkflowOptions{
		ID:        deliveryPackage.ID,
		TaskQueue: d.PackageDeliveryTaskQueueName,
	}

	workflowInput := workflow.PackageDeliveryWorkflowParams{
		DeliveryPackage: &model.DeliveryPackage{
			CustomerEmail:   deliveryPackage.CustomerEmail,
			DeliveryAddress: deliveryPackage.DeliveryAddress,
		},
	}

	_, err := d.TemporalClient.ExecuteWorkflow(context.Background(), wo, workflow.PackageDeliveryWorkflowName, workflowInput)
	if err != nil {
		d.Logger.Error("temporal client execute workflow", zap.Error(err))
		return err
	}

	return nil
}
