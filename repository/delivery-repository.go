package repository

import (
	"fmt"
	"go-test/internal/model"
	"go.uber.org/zap"
)

func (r *Repository) CreatePackageDelivery(payload *model.DeliveryPackage) (*model.DeliveryPackage, error) {
	deliveryPackage := &model.DeliveryPackage{
		ID:              payload.ID,
		CustomerEmail:   payload.CustomerEmail,
		DeliveryAddress: payload.DeliveryAddress,
	}

	if err := r.Connection.Create(deliveryPackage).Error; err != nil {
		r.Logger.Error("Failed to create delivery package", zap.String("package_id", payload.ID), zap.Error(err))
		return nil, fmt.Errorf("failed to create package delivery: %w", err)
	}

	r.Logger.Info("Successfully created delivery package", zap.String("package_id", payload.ID))

	return deliveryPackage, nil
}
