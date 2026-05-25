package workflow

import (
	"cmp"
	"context"
	"slices"
)

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

func (e *Executor[T]) Execute(ctx context.Context, step WorkflowStep, args any) (WorkflowResult, error) {
	stepHandler := e.WfExec.stepOperators[step]
	getStepResult := e.WfExec.stepResults[step]

	stepResult, err := getStepResult(ctx, args)

	skipResults := []WorkflowResult{Skip, Succeed}
	if slices.Contains(skipResults, stepResult) {
		return stepResult, err
	}

	result, err := stepHandler(ctx, args)

	var msgErr string
	if err != nil || result == Failed {
		msgErr = cmp.Or(err.Error(), string(result))
		e.WfExec.SetStepResult(step, WorkflowProcess{
			Status: result,
			Response: msgErr,
		})
	} else {
		e.WfExec.SetStepResult(step, WorkflowProcess{
			Status: result,
			Response: result,
		})
	}

	return result, err
}
