package workflow

import (
	"slices"
)

type stepHandler func(args ...any) (WorkflowResult, error)
type ExecuteFunc func(executor *executor) (WorkflowResult, error)

type executor struct {
	wfExec *WorkflowExecutor
}

func NewExecutor(wfExec *WorkflowExecutor) *executor {
	return &executor{
		wfExec: wfExec,
	}
}

func (e *executor) Execute(step string, args ...any) (WorkflowResult, error) {
	stepHandler := e.wfExec.stepOperators[step]
	getStepResult := e.wfExec.stepResults[step]

	stepResult, err := getStepResult(args...)

	skipResults := []WorkflowResult{Failed, Skip, Succeed}
	if slices.Contains(skipResults, stepResult) {
		return stepResult, err
	}

	result, err := stepHandler(args...)

	return result, err
}
