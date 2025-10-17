package workflow

import (
	"cmp"
	"context"
	"go-base/pkg"
	"go-base/pkg/container"
	"go-base/pkg/datasource/postgres/repository"
	"slices"
	"time"
)

type WorkflowProps struct {
	Name       string
	Payload    any
	MaxAttempt int
	Schedule   string
}

type WorkflowExecutor struct {
	container *container.Container
	cron      *pkg.Cronjob
	repo      *WorkflowRepository

	// Executor to run the workflow
	executor    *executor
	ExecuteFunc ExecuteFunc

	// Workflow properties
	props WorkflowProps

	stepOperators map[string]stepHandler
	stepResults   map[string]stepHandler
}

func NewWorkflowExecutor(
	ctn *container.Container,
	props WorkflowProps,
	repo *repository.BaseRepository,
	execFn ExecuteFunc,
) *WorkflowExecutor {
	wfExec := &WorkflowExecutor{
		cron:          ctn.NewCron(),
		props:         props,
		container:     ctn,
		repo:          newWorkflowRepository(repo),
		stepOperators: make(map[string]stepHandler),
		stepResults:   make(map[string]stepHandler),
		ExecuteFunc:   execFn,
	}

	wfExec.executor = NewExecutor(wfExec)

	return wfExec
}

func (w *WorkflowExecutor) Init() {
	w.container.AddCronJob(w.props.Schedule, w.props.Name, w.crawlWorkflow)
}

func (w *WorkflowExecutor) crawlWorkflow() {
	ctx := context.Background()
	wfs, err := w.getRerunWorkflows(ctx)

	if err != nil {
		w.container.Logger.Error("Failed to get rerun workflows:", err)
	}

	for _, wf := range wfs {
		go w.runWorkflow(ctx, &wf)
	}
}

func (w *WorkflowExecutor) isFinished(workflow *Workflow) bool {
	return slices.Contains([]WorkflowResult{Completed, Failed}, workflow.Status)
}

func (w *WorkflowExecutor) runWorkflow(ctx context.Context, wf *Workflow) bool {
	if wf == nil {
		return false
	}

	lock, errObtainLock := w.container.Locker.Obtain(ctx, wf.ID, WF_DEFAULT_TIMEOUT, nil)
	defer lock.Release(ctx)

	if errObtainLock != nil {
		return false
	}

	var currentAttempt int
	startTime := time.Now()
	if wf.CurrentAttempt == wf.MaxAttempts {
		currentAttempt = wf.CurrentAttempt
	} else {
		currentAttempt = wf.CurrentAttempt + 1
	}

	if wf.Finished {
		return true
	}

	_, err := w.ExecuteFunc(w.executor)

	if err != nil {
		w.container.Logger.Error("Workflow execution failed:", err)
		wf.ProcessResults = err.Error()
	}

	duration := time.Since(startTime)
	wf.StartedTime = startTime
	wf.FinishedTime = startTime.Add(duration)
	wf.CurrentAttempt = currentAttempt
	wf.Finished = cmp.Or(wf.CurrentAttempt >= w.props.MaxAttempt, w.isFinished(wf))

	w.repo.baseRepo.Update(ctx, wf)

	return true
}
