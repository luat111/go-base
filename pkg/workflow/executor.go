package workflow

import (
	"cmp"
	"context"
	"slices"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
)

const tracerName = "go-base/workflow"

type (
	ExecuteFunc[T ~struct{ Workflow }] func(ctx context.Context, executor *Executor[T], repo IWorkflowRepository[T]) (WorkflowResult, error)

	Executor[T ~struct{ Workflow }] struct {
		WfExec *WorkflowExecutor[T]
	}
)

func NewExecutor[T ~struct{ Workflow }](wfExec *WorkflowExecutor[T]) *Executor[T] {
	return &Executor[T]{
		WfExec: wfExec,
	}
}

// Execute runs a single workflow step wrapped in an OpenTelemetry span.
// The span name follows the pattern "workflow.step <stepName>" and the
// result / error are recorded on the span before it ends.
func (e *Executor[T]) Execute(ctx context.Context, step WorkflowStep, args any) (WorkflowResult, error) {
	tracer := otel.Tracer(tracerName)
	ctx, span := tracer.Start(ctx, "workflow.step "+string(step))
	defer span.End()

	span.SetAttributes(
		attribute.String("workflow.step", string(step)),
		attribute.String("workflow.name", e.WfExec.props.Name),
	)

	stepHandler := e.WfExec.stepOperators[step]
	getStepResult := e.WfExec.stepResults[step]

	stepResult, err := getStepResult(ctx, args)

	skipResults := []WorkflowResult{Skip, Succeed}
	if slices.Contains(skipResults, stepResult) {
		span.SetAttributes(attribute.String("workflow.step.result", string(stepResult)))
		return stepResult, err
	}

	result, err := stepHandler(ctx, args)

	span.SetAttributes(attribute.String("workflow.step.result", string(result)))

	var msgErr string
	if err != nil || result == Failed {
		msgErr = cmp.Or(err.Error(), string(result))
		span.RecordError(err)
		span.SetStatus(codes.Error, msgErr)
		e.WfExec.SetStepResult(step, WorkflowProcess{
			Status:   result,
			Response: msgErr,
		})
	} else {
		span.SetStatus(codes.Ok, "")
		e.WfExec.SetStepResult(step, WorkflowProcess{
			Status:   result,
			Response: result,
		})
	}

	return result, err
}
