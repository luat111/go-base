package workflow

import (
	"context"
	"go-base/pkg/common/types"
	"go-base/pkg/datasource/postgres/repository"
)

type WorkflowRepository struct {
	baseRepo *repository.BaseRepository
	Model    any
}

func NewWorkflowRepository(repo *repository.BaseRepository) *WorkflowRepository {
	return &WorkflowRepository{baseRepo: repo}
}

func (w *WorkflowExecutor) getRerunWorkflows(ctx context.Context) ([]Workflow, error) {
	var workflows []Workflow

	pendingStatus := []WorkflowResult{New, Processing}

	err := w.repo.baseRepo.DB.WithContext(ctx).Where("status IN ? AND finished = false", pendingStatus).Find(&workflows).Error

	return workflows, err
}

func (w *WorkflowExecutor) createWorkflow(ctx context.Context) error {
	var wf Workflow

	wf.Status = New
	wf.CurrentAttempt = 1
	wf.ProcessResults = types.JSONB{}
	wf.WorkflowName = w.props.Name
	wf.MaxAttempts = w.props.MaxAttempt
	wf.Payload = w.Payload

	return w.repo.baseRepo.Create(ctx, &wf)
}
