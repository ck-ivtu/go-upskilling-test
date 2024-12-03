package activities

import (
	"context"
	"go-test/internal/model"
	"go-test/repository"
	"go.temporal.io/sdk/activity"
	"go.uber.org/zap"
)

const SaveDeliveryActivityName = "save-delivery-activity"

type SaveDelivery struct {
	Repo   *repository.Repository
	Logger *zap.Logger
}

type SaveDeliveryInput struct {
	ID              string
	DeliveryPackage *model.DeliveryPackage
}

func NewSaveDelivery(repo *repository.Repository, logger *zap.Logger) *SaveDelivery {
	return &SaveDelivery{Repo: repo, Logger: logger}
}

func (s *SaveDelivery) SaveDeliveryActivity(ctx context.Context, params *SaveDeliveryInput) (*model.DeliveryPackage, error) {
	attempt := int(activity.GetInfo(ctx).Attempt)

	s.Logger.Info("Starting save delivery activity", zap.Int("attempt", attempt))

	pack, err := s.Repo.CreatePackageDelivery(&model.DeliveryPackage{
		ID:              params.ID,
		DeliveryAddress: params.DeliveryPackage.DeliveryAddress,
		CustomerEmail:   params.DeliveryPackage.CustomerEmail,
	})
	if err != nil {
		s.Logger.Error("Failed to save delivery package", zap.Error(err), zap.String("packageId", params.ID))
		return nil, err
	}

	s.Logger.Info("Successfully saved delivery package", zap.String("packageId", params.ID))

	return pack, nil
}
