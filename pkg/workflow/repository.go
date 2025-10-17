package workflow

import (
	"gorm.io/gorm"
)

type WorkflowRepository struct {
	DB *gorm.DB
}

func (r *WorkflowRepository) Create(workflow *Workflow) error {
	return r.DB.Create(workflow).Error
}

func (r *WorkflowRepository) Update(workflow *Workflow) error {
	return r.DB.Save(workflow).Error
}

func (r *WorkflowRepository) GetPendingWorkflows() ([]Workflow, error) {
	var workflows []Workflow
	err := r.DB.Where("status = ?", "NEW").Find(&workflows).Error
	return workflows, err
}
