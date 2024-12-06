package controllers

import (
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	_ "go-test/docs"
	"go-test/internal/controllers/packages"
	"go-test/internal/events"
	"go.temporal.io/sdk/client"
	"go.uber.org/zap"
)

const ApiV1Path = "/api/v1"
const PackagesPath = "/packages"

func InitializeRoutes(logger *zap.Logger, temporalClient client.Client, r *gin.Engine, ep *events.EventProducer) *gin.Engine {
	createPackageController := packages.RegisterCreatePackageController(logger, temporalClient, ep)
	getPackageController := packages.RegisterGetPackageController(logger, temporalClient)
	confirmPackageController := packages.RegisterConfirmPackageController(logger, temporalClient)

	apiV1Group := r.Group(ApiV1Path)

	packagesGroup := apiV1Group.Group(PackagesPath)
	packagesGroup.POST("/", createPackageController.CreatePackage)
	packagesGroup.GET("/:id", getPackageController.GetPackage)
	packagesGroup.POST("/:id/confirm", confirmPackageController.ConfirmPackage)

	apiV1Group.GET("/docs/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	return r
}
