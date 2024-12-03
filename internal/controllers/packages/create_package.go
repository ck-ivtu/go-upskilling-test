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

type CreatePackageController struct {
	Logger                       *zap.Logger
	TemporalClient               client.Client
	PackageDeliveryTaskQueueName string
}

func RegisterCreatePackageController(logger *zap.Logger, temporalClient client.Client) *CreatePackageController {
	return &CreatePackageController{
		Logger:                       logger,
		TemporalClient:               temporalClient,
		PackageDeliveryTaskQueueName: workflow.PackageDeliveryTaskQueueName,
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

	wo := client.StartWorkflowOptions{
		TaskQueue: c.PackageDeliveryTaskQueueName,
	}

	workflowInput := workflow.PackageDeliveryWorkflowParams{
		DeliveryPackage: &model.DeliveryPackage{
			CustomerEmail:   req.CustomerEmail,
			DeliveryAddress: req.DeliveryAddress,
		},
	}

	run, err := c.TemporalClient.ExecuteWorkflow(context.Background(), wo, workflow.PackageDeliveryWorkflowName, workflowInput)
	if err != nil {
		c.Logger.Error("temporal client execute workflow", zap.Error(err))
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to execute order workflow"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"Package ID": run.GetID()})
}
