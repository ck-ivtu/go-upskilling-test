package packages

import (
	"context"
	"github.com/gin-gonic/gin"
	"go-test/internal/model"
	_ "go-test/internal/model"
	"go-test/internal/workflow"
	"go.temporal.io/sdk/client"
	"go.uber.org/zap"
	"net/http"
)

type ConfirmPackageResponse struct {
	Status model.PackageDeliveryState `json:"status"`
}

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

// ConfirmPackage godoc
// @Summary      Confirm package delivery
// @Description  Confirm the delivery of a package
// @Tags         packages
// @Accept       json
// @Produce      json
// @Param        id path string true "Package ID"
// @Success      200 {object} ConfirmPackageResponse "Confirmation status"
// @Failure      400 {object} model.HttpErrorResponse "Invalid input data"
// @Failure      502 {object} model.HttpErrorResponse "Unable to confirm package"
// @Router       /api/v1/packages/{id}/confirm [post]
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

	ctx.JSON(http.StatusOK, &ConfirmPackageResponse{Status: model.PackageDeliveryConfirmed})
}
