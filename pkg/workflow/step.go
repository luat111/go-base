package workflow

import (
	"context"
	"errors"
)

type (
	WorkflowStep string

	StepHandler func(ctx context.Context, args any) (WorkflowResult, error)

	WorkflowProcess struct {
		Status   WorkflowResult
		Response any
	}
)

func (w *WorkflowExecutor[T]) SetStepOperators(stepOperators map[WorkflowStep]StepHandler) {
	w.stepOperators = stepOperators
}

func (w *WorkflowExecutor[T]) SetStepResults(stepResults map[WorkflowStep]StepHandler) {
	w.stepResults = stepResults
}

func (w *WorkflowExecutor[T]) GetStepResult(step WorkflowStep) (WorkflowProcess, WorkflowResult) {
	result := w.ProcessResults[string(step)]

	cvrt, ok := result.(WorkflowProcess)
	if !ok {
		return cvrt, Failed
	}

	status := WorkflowResult(cvrt.Status)

	return cvrt, status
}

func (w *WorkflowExecutor[T]) SetStepResult(step WorkflowStep, result WorkflowProcess) {
	w.mu.Lock()
	defer w.mu.Unlock()

	w.ProcessResults[string(step)] = result
}

func (w *WorkflowExecutor[T]) OverrideStepResponse(step WorkflowStep, response any) error {
	w.mu.Lock()
	defer w.mu.Unlock()

	result := w.ProcessResults[string(step)]
	convert, ok := result.(*WorkflowProcess)
	if !ok {
		w.logger.Error("Can not convert step result")
		return errors.New("Can not convert step result")
	}

	convert.Response = response
	w.ProcessResults[string(step)] = result

	return nil
}
