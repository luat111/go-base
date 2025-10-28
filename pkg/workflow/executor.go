package workflow

import (
	"slices"
)

type ExecuteFunc func(executor *Executor) (WorkflowResult, error)

type Executor struct {
	wfExec *WorkflowExecutor
}

func NewExecutor(wfExec *WorkflowExecutor) *Executor {
	return &Executor{
		wfExec: wfExec,
	}
}

func (e *Executor) Execute(step WorkflowStep, args any) (WorkflowResult, error) {
	stepHandler := e.wfExec.stepOperators[step]
	getStepResult := e.wfExec.stepResults[step]

	stepResult, err := getStepResult(args)

	skipResults := []WorkflowResult{Failed, Skip, Succeed}
	if slices.Contains(skipResults, stepResult) {
		return stepResult, err
	}

	return stepHandler(args)
}
