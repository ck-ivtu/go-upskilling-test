package controllers

import (
	"github.com/gin-gonic/gin"
	"go-test/internal/controllers/packages"
	"go-test/internal/events"
	"go.temporal.io/sdk/client"
	"go.uber.org/zap"
)

const PackagesPath = "/packages"

func InitializeRoutes(logger *zap.Logger, temporalClient client.Client, r *gin.Engine, ep *events.EventProducer) *gin.Engine {
	createPackageController := packages.RegisterCreatePackageController(logger, temporalClient, ep)
	getPackageController := packages.RegisterGetPackageController(logger, temporalClient)
	confirmPackageController := packages.RegisterConfirmPackageController(logger, temporalClient)

	packagesGroup := r.Group(PackagesPath)
	packagesGroup.POST("/", createPackageController.CreatePackage)
	packagesGroup.GET("/:id", getPackageController.GetPackage)
	packagesGroup.POST("/:id/confirm", confirmPackageController.ConfirmPackage)

	return r
}
