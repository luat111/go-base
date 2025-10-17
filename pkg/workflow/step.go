package workflow

func (w *WorkflowExecutor) SetStepOperators(stepOperators map[string]stepHandler) {
	w.stepOperators = stepOperators
}

func (w *WorkflowExecutor) SetStepResults(stepResults map[string]stepHandler) {
	w.stepResults = stepResults
}
