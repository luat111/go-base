package workflow

type WorkflowStep string
type StepHandler func(args any) (WorkflowResult, error)

func (w *WorkflowExecutor) SetStepOperators(stepOperators map[WorkflowStep]StepHandler) {
	w.stepOperators = stepOperators
}

func (w *WorkflowExecutor) SetStepResults(stepResults map[WorkflowStep]StepHandler) {
	w.stepResults = stepResults
}
