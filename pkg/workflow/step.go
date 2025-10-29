package workflow

import "context"

type WorkflowStep string
type StepHandler func(ctx context.Context, args any) (WorkflowResult, error)

func (w *WorkflowExecutor[T]) SetStepOperators(stepOperators map[WorkflowStep]StepHandler) {
	w.stepOperators = stepOperators
}

func (w *WorkflowExecutor[T]) SetStepResults(stepResults map[WorkflowStep]StepHandler) {
	w.stepResults = stepResults
}
