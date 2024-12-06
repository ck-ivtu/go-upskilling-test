package packages

import (
	"context"
	"encoding/json"
	"github.com/gin-gonic/gin"
	_ "go-test/internal/model"
	"go-test/internal/workflow"
	"go.temporal.io/sdk/client"
	"go.uber.org/zap"
	"net/http"
)

type GetPackageController struct {
	Logger         *zap.Logger
	TemporalClient client.Client
}

func RegisterGetPackageController(logger *zap.Logger, temporalClient client.Client) *GetPackageController {
	return &GetPackageController{
		Logger:         logger,
		TemporalClient: temporalClient,
	}
}

// GetPackage godoc
// @Summary      Get package details
// @Description  Get details of a specific package delivery
// @Tags         packages
// @Accept       json
// @Produce      json
// @Param        id path string true "Package ID"
// @Success      200 {object} model.DeliveryPackage
// @Failure      400 {object} model.HttpErrorResponse "Invalid input data"
// @Failure      404 {object} model.HttpErrorResponse "Package not found"
// @Router       /api/v1/packages/{id} [get]
func (c *GetPackageController) GetPackage(ctx *gin.Context) {
	packageId := ctx.Param("id")
	if packageId == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Package ID is required"})
		return
	}

	wf, err := c.TemporalClient.QueryWorkflow(context.Background(), packageId, "", workflow.PackageDeliveryStateQuery)
	if err != nil {
		c.Logger.Error("Error querying Temporal client", zap.String("packageId", packageId), zap.Error(err))
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to query workflow"})
		return
	}

	var queryResult interface{}
	err = wf.Get(&queryResult)
	if err := wf.Get(&queryResult); err != nil {
		c.Logger.Error("Error getting query result", zap.String("packageId", packageId), zap.Error(err))
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve query result"})
		return
	}

	var resultData gin.H
	switch value := queryResult.(type) {
	case string:
		resultData = gin.H{"status": value}
	default:
		r, err := json.Marshal(value)
		if err != nil {
			c.Logger.Error("Error marshaling query result", zap.String("packageId", packageId), zap.Error(err))
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to process query result"})
			return
		}

		var workflowResult workflow.PackageDeliveryWorkflowResult
		if err := json.Unmarshal(r, &workflowResult); err != nil {
			c.Logger.Error("Error unmarshaling query result", zap.String("packageId", packageId), zap.Error(err))
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to decode workflow result"})
			return
		}

		resultData = gin.H{"status": workflowResult.Status}
	}

	ctx.JSON(http.StatusOK, resultData)
}
