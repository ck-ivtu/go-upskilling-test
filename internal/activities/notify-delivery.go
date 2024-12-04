package activities

import (
	"context"
	"go-test/internal/adapters"
	"go-test/internal/model"
	"go.temporal.io/sdk/activity"
	"go.uber.org/zap"
)

const NotifyDeliveryActivityName = "notify-delivery-activity"

type NotifyDelivery struct {
	Logger *zap.Logger
}

type NotifyDeliveryInput struct {
	ID              string
	DeliveryPackage *model.DeliveryPackage
}

func NewNotifyDelivery(logger *zap.Logger) *NotifyDelivery {
	return &NotifyDelivery{
		Logger: logger,
	}
}

func (n *NotifyDelivery) NotifyDeliveryActivity(ctx context.Context, input *NotifyDeliveryInput) error {
	attempt := int(activity.GetInfo(ctx).Attempt)

	n.Logger.Info("Starting notify delivery activity", zap.Int("attempt", attempt))

	// TODO add config
	notifyDeliveryClient := adapters.NewNotifyDeliveryClient("3af31544-ce24-4f48-b563-f5a8ba38656e", n.Logger)

	err := notifyDeliveryClient.Notify(ctx, *input.DeliveryPackage)

	if err != nil {
		n.Logger.Error("Failed to notify delivery activity", zap.Error(err))
		return err
	}

	return nil
}
