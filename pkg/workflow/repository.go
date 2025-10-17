package workflow

import (
	"context"
	"go-base/pkg/datasource/postgres/repository"
)

type WorkflowRepository struct {
	baseRepo *repository.BaseRepository
}

func newWorkflowRepository(repo *repository.BaseRepository) *WorkflowRepository {
	return &WorkflowRepository{baseRepo: repo}
}

func (w *WorkflowExecutor) getRerunWorkflows(ctx context.Context) ([]Workflow, error) {
	var workflows []Workflow

	pendingStatus := []WorkflowResult{New, Processing}

	err := w.repo.baseRepo.DB.WithContext(ctx).Where("status IN ?", pendingStatus).Find(&workflows).Error

	return workflows, err
}

func (w *WorkflowExecutor) createWorkflow(ctx context.Context, payload *Workflow) error {
	payload.Status = New
	payload.CurrentAttempt = 1
	payload.ProcessResults = ""
	payload.WorkflowName = w.props.Name
	payload.MaxAttempts = w.props.MaxAttempt

	return w.repo.baseRepo.Create(ctx, payload)
}
