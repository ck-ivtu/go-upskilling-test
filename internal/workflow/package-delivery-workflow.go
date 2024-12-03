package workflow

import (
	"go-test/internal/activities"
	"go-test/internal/model"
	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
	"go.uber.org/zap"
	"time"
)

func NewPackageDeliveryWorkflowConfig(logger *zap.Logger) *PackageDeliveryWorkflowConfig {
	return &PackageDeliveryWorkflowConfig{
		Logger: logger,
	}
}

func newPackageDeliveryWorkflow(ctx workflow.Context, params *PackageDeliveryWorkflowParams) *PackageDeliveryWorkflow {
	return &PackageDeliveryWorkflow{
		Ctx:            ctx,
		State:          NewPackageDeliveryWorkflowState(),
		Package:        params.DeliveryPackage,
		WorkflowResult: &PackageDeliveryWorkflowResult{Status: model.PackageDeliveryInProgress},
	}
}

func (c *PackageDeliveryWorkflowConfig) PackageDeliveryWorkflow(
	ctx workflow.Context,
	params *PackageDeliveryWorkflowParams,
) (workflowResult *PackageDeliveryWorkflowResult, err error) {
	w := newPackageDeliveryWorkflow(ctx, params)

	c.Logger.Info("Starting package delivery workflow", zap.String("workflowId", workflow.GetInfo(ctx).WorkflowExecution.ID))

	workflow.Go(ctx, func(goCtx workflow.Context) {
		sel := workflow.NewSelector(goCtx)

		confirm := workflow.GetSignalChannel(goCtx, PackageDeliverySignalConfirm)

		sel.AddReceive(confirm, func(ch workflow.ReceiveChannel, more bool) {
			ch.Receive(goCtx, w.State.DeliveryConfirmed)
		})

		for {
			sel.Select(goCtx)
		}

		// TODO setup timers
	})

	if err := workflow.SetQueryHandler(ctx, PackageDeliveryStateQuery, func() (PackageDeliveryWorkflowResult, error) {
		return *w.WorkflowResult, nil
	}); err != nil {
		w.WorkflowResult.Status = model.PackageDeliveryErrored

		return w.WorkflowResult, err
	}

	workflow.Await(ctx, w.State.ShouldHandlePackageDeliveryConfirm)

	w.WorkflowResult.Status = model.PackageDeliveryConfirmed

	activityOptions := workflow.ActivityOptions{
		StartToCloseTimeout: time.Minute,
		RetryPolicy: &temporal.RetryPolicy{
			MaximumAttempts: 3,
		},
	}

	activityCtx := workflow.WithActivityOptions(ctx, activityOptions)

	err = workflow.ExecuteActivity(
		activityCtx,
		activities.SaveDeliveryActivityName,
		&activities.SaveDeliveryInput{
			ID: workflow.GetInfo(ctx).WorkflowExecution.ID,
			DeliveryPackage: &model.DeliveryPackage{
				DeliveryAddress: params.DeliveryPackage.DeliveryAddress,
				CustomerEmail:   params.DeliveryPackage.CustomerEmail,
			},
		},
	).Get(ctx, nil)

	if err != nil {
		w.WorkflowResult.Status = model.PackageDeliveryErrored
		c.Logger.Error("Failed to save delivery activity", zap.Error(err))

		return w.WorkflowResult, err
	}

	w.WorkflowResult.Status = model.PackageDeliverySaved

	// TODO trigger webhook activity

	return w.WorkflowResult, nil
}
