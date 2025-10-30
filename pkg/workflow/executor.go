package workflow

import (
	"cmp"
	"context"
	"slices"
)

type ExecuteFunc[T ~struct{ Workflow }] func(ctx context.Context, executor *Executor[T], repo IWorkflowRepository[T]) (WorkflowResult, error)

type Executor[T ~struct{ Workflow }] struct {
	wfExec *WorkflowExecutor[T]
}

func NewExecutor[T ~struct{ Workflow }](wfExec *WorkflowExecutor[T]) *Executor[T] {
	return &Executor[T]{
		wfExec: wfExec,
	}
}

func (e *Executor[T]) Execute(ctx context.Context, step WorkflowStep, args any) (WorkflowResult, error) {
	stepHandler := e.wfExec.stepOperators[step]
	getStepResult := e.wfExec.stepResults[step]

	stepResult, err := getStepResult(ctx, args)

	skipResults := []WorkflowResult{Skip, Succeed}
	if slices.Contains(skipResults, stepResult) {
		return stepResult, err
	}

	result, err := stepHandler(ctx, args)

	var msgErr string
	if err != nil || result == Failed {
		msgErr = cmp.Or(err.Error(), string(result))
		e.wfExec.SetStepResult(step, msgErr)
	} else {
		e.wfExec.SetStepResult(step, result)
	}

	return result, err
}
