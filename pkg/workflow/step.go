package workflow

import (
	"context"
)

type WorkflowStep string

type StepHandler func(ctx context.Context, args any) (WorkflowResult, error)

func (w *WorkflowExecutor[T]) SetStepOperators(stepOperators map[WorkflowStep]StepHandler) {
	w.stepOperators = stepOperators
}

func (w *WorkflowExecutor[T]) SetStepResults(stepResults map[WorkflowStep]StepHandler) {
	w.stepResults = stepResults
}

func (w *WorkflowExecutor[T]) GetStepResult(step WorkflowStep) WorkflowResult {
	result := w.ProcessResults[string(step)]

	if res, ok := result.(WorkflowResult); ok {
		return res
	}

	return Failed
}

func (w *WorkflowExecutor[T]) SetStepResult(step WorkflowStep, result any) {
	w.mu.Lock()
	defer w.mu.Unlock()

	w.ProcessResults[string(step)] = result
}
