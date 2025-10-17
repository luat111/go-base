package workflow

import (
	"encoding/json"
)

type WorkflowExecutor struct {
	Repo *WorkflowRepository
}

func (e *WorkflowExecutor) Execute(workflow *Workflow) error {
	// Define steps
	steps := []func(payload map[string]interface{}) (string, error){}

	var payload map[string]interface{}
	json.Unmarshal([]byte(workflow.Payload), &payload)

	for _, step := range steps {
		result, err := step(payload)
		if err != nil {
			workflow.CurrentAttempt++
			if workflow.CurrentAttempt >= workflow.MaxAttempts {
				workflow.Status = "FAILED"
				e.Repo.Update(workflow)
				return err
			}
			continue
		}
		workflow.ProcessResults += result + ";"
	}

	workflow.Status = "COMPLETED"
	return e.Repo.Update(workflow)
}
