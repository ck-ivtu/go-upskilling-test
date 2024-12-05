package packages

import (
	"encoding/json"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go-test/internal/events"
	"go-test/internal/model"
	"go-test/internal/workflow"
	"go.temporal.io/sdk/client"
	"go.uber.org/zap"
	"net/http"
)

type CreatePackageController struct {
	Logger                       *zap.Logger
	TemporalClient               client.Client
	PackageDeliveryTaskQueueName string
	EventProducer                *events.EventProducer
}

func RegisterCreatePackageController(
	logger *zap.Logger,
	temporalClient client.Client,
	eventProducer *events.EventProducer,
) *CreatePackageController {
	return &CreatePackageController{
		Logger:                       logger,
		TemporalClient:               temporalClient,
		PackageDeliveryTaskQueueName: workflow.PackageDeliveryTaskQueueName,
		EventProducer:                eventProducer,
	}
}

type CreatePackageRequest struct {
	CustomerEmail   string `json:"customer_email" binding:"required,email"`
	DeliveryAddress string `json:"delivery_address" binding:"required"`
}

func (c *CreatePackageController) CreatePackage(ctx *gin.Context) {
	var req CreatePackageRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input data"})
		return
	}

	if req.CustomerEmail == "" || req.DeliveryAddress == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "All fields are required"})
		return
	}

	deliveryTrackingId := uuid.New().String()

	deliveryPackage := &model.DeliveryPackage{
		ID:              deliveryTrackingId,
		CustomerEmail:   req.CustomerEmail,
		DeliveryAddress: req.DeliveryAddress,
	}

	event, err := json.Marshal(deliveryPackage)
	if err != nil {
		c.Logger.Error("failed to marshal delivery package", zap.Error(err))
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create delivery package, try again"})
		return
	}

	err = c.EventProducer.SendEvent(string(event))
	if err != nil {
		c.Logger.Error("failed to create delivery package", zap.Error(err))
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create delivery package, try again"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"Package ID": deliveryTrackingId})
}
