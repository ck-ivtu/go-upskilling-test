package packages

import (
	"context"
	"github.com/gin-gonic/gin"
	"go-test/internal/model"
	"go-test/internal/workflow"
	"go.temporal.io/sdk/client"
	"go.uber.org/zap"
	"net/http"
)

type ConfirmPackageController struct {
	Logger                       *zap.Logger
	TemporalClient               client.Client
	PackageDeliveryTaskQueueName string
}

func RegisterConfirmPackageController(logger *zap.Logger, temporalClient client.Client) *ConfirmPackageController {
	return &ConfirmPackageController{
		Logger:                       logger,
		TemporalClient:               temporalClient,
		PackageDeliveryTaskQueueName: workflow.PackageDeliveryTaskQueueName,
	}
}

func (c *ConfirmPackageController) ConfirmPackage(ctx *gin.Context) {
	packageId := ctx.Param("id")

	if packageId == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Package ID is required"})
		return
	}

	err := c.TemporalClient.SignalWorkflow(
		context.Background(),
		packageId,
		"",
		workflow.PackageDeliverySignalConfirm,
		model.DeliveryPackageWorkflowStatus{RequestReceived: true},
	)
	if err != nil {
		c.Logger.Error("Unable to signal workflow", zap.Error(err))

		ctx.JSON(http.StatusBadGateway, gin.H{"error": "Unable to confirm package, package is already delivered"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"status": "confirmed"})
}
