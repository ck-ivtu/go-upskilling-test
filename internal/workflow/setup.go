package workflow

import (
	"go-test/internal/activities"
	"go-test/repository"
	"go.temporal.io/sdk/activity"
	"go.temporal.io/sdk/worker"
	"go.temporal.io/sdk/workflow"
	"go.uber.org/zap"
)

func SetupWorkflow(w worker.Worker, r *repository.Repository, logger *zap.Logger) {
	w.RegisterWorkflowWithOptions(NewPackageDeliveryWorkflowConfig(logger).PackageDeliveryWorkflow, workflow.RegisterOptions{
		Name: PackageDeliveryWorkflowName,
	})

	SetupActivities(w.RegisterActivityWithOptions, r, logger)
}

func SetupActivities(RegisterActivityWithOptions func(a interface{}, options activity.RegisterOptions), r *repository.Repository, logger *zap.Logger) {
	RegisterActivityWithOptions(activities.NewSaveDelivery(r, logger).SaveDeliveryActivity, activity.RegisterOptions{
		Name: activities.SaveDeliveryActivityName,
	})
}
