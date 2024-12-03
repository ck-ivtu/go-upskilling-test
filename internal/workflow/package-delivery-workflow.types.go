package workflow

import (
	"go-test/internal/model"
	"go.temporal.io/sdk/workflow"
	"go.uber.org/zap"
)

const (
	PackageDeliveryWorkflowName  = "package-delivery-workflow"
	PackageDeliveryTaskQueueName = "package-delivery-task-queue"
)

const (
	PackageDeliverySignalConfirm = "confirm"
	PackageDeliveryStateQuery    = "current-state"
)

type PackageDeliveryWorkflowConfig struct {
	Logger *zap.Logger
}

type PackageDeliveryWorkflowParams struct {
	DeliveryPackage *model.DeliveryPackage
}

type PackageDeliveryWorkflowResult struct {
	Status model.PackageDeliveryState `json:"status"`
}

type PackageDeliveryWorkflowState struct {
	DeliveryConfirmed *model.DeliveryPackageWorkflowStatus

	Pending   bool
	Completed bool
}

func (s *PackageDeliveryWorkflowState) ShouldHandlePackageDeliveryConfirm() bool {
	return s.DeliveryConfirmed.ShouldHandle()
}

func NewPackageDeliveryWorkflowState() *PackageDeliveryWorkflowState {
	return &PackageDeliveryWorkflowState{
		DeliveryConfirmed: &model.DeliveryPackageWorkflowStatus{},
		Pending:           true,
	}
}

type PackageDeliveryWorkflow struct {
	Ctx            workflow.Context
	State          *PackageDeliveryWorkflowState
	Package        *model.DeliveryPackage
	WorkflowResult *PackageDeliveryWorkflowResult
}
