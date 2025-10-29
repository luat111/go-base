package workflow

import (
	"context"
	"go-base/pkg/datasource/postgres/repository"
)

type IWorkflowRepository[T ~struct{ *Workflow }] interface {
	GetRerunWorkflows(ctx context.Context) ([]T, error)
	CreateWorkflow(ctx context.Context, payload any) error
	Update(ctx context.Context, entity any) error
}

type WorkflowRepository[T ~struct{ *Workflow }] struct {
	baseRepo *repository.BaseRepository
}

func NewWorkflowRepository[T ~struct{ *Workflow }](repo *repository.BaseRepository) IWorkflowRepository[T] {
	return &WorkflowRepository[T]{baseRepo: repo}
}

func (w *WorkflowRepository[T]) GetRerunWorkflows(ctx context.Context) ([]T, error) {
	var workflows []T

	pendingStatus := []WorkflowResult{New, Processing}

	err := w.baseRepo.DB.WithContext(ctx).Where("status IN ? AND finished = false", pendingStatus).Find(&workflows).Error

	return workflows, err
}

func (w *WorkflowRepository[T]) CreateWorkflow(ctx context.Context, payload any) error {
	return w.baseRepo.Create(ctx, payload)
}

func (w *WorkflowRepository[T]) Update(ctx context.Context, payload any) error {
	return w.baseRepo.Create(ctx, payload)
}
