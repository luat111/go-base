package workflow

import (
	"cmp"
	"context"
	"go-base/pkg"
	"go-base/pkg/common"
	"go-base/pkg/common/types"
	"go-base/pkg/container"
	"go-base/pkg/logger"
	"slices"
	"sync"
	"time"
)

type WorkflowProps struct {
	Name       string
	Payload    types.JSONB
	MaxAttempt int
	Schedule   string
}

type WorkflowExecutor struct {
	container *container.Container
	cron      *pkg.Cronjob
	repo      *WorkflowRepository

	// Executor to run the workflow
	Executor    *Executor
	ExecuteFunc ExecuteFunc

	// Workflow properties
	props       WorkflowProps
	retryConfig RetryConfig

	// Workflow data
	ProcessResults types.JSONB
	Payload        types.JSONB

	stepOperators map[WorkflowStep]StepHandler
	stepResults   map[WorkflowStep]StepHandler

	logger logger.ILogger

	mu sync.Mutex
}

func NewWorkflowExecutor(
	ctn *container.Container,
	props WorkflowProps,
	repo *WorkflowRepository,
	execFn ExecuteFunc,
	retryConfig RetryConfig,
) *WorkflowExecutor {
	logger := logger.NewLogger(common.WorkflowPrefix)
	wfExec := &WorkflowExecutor{
		cron:          ctn.NewCron(),
		props:         props,
		container:     ctn,
		repo:          repo,
		stepOperators: make(map[WorkflowStep]StepHandler),
		stepResults:   make(map[WorkflowStep]StepHandler),
		ExecuteFunc:   execFn,
		retryConfig:   retryConfig,
		logger:        logger,
	}

	wfExec.Executor = NewExecutor(wfExec)

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

func (w *WorkflowExecutor) SetProcessResults(processResults map[string]any) {
	w.mu.Lock()
	defer w.mu.Unlock()

	w.ProcessResults = processResults
}

func (w *WorkflowExecutor) SetPayload(payload map[string]any) {
	w.mu.Lock()
	defer w.mu.Unlock()

	w.Payload = payload
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

	// Map for execute
	w.mu.Lock()
	w.ProcessResults, w.Payload = wf.ProcessResults, wf.Payload
	w.mu.Unlock()

	_, err := w.ExecuteFunc(w.Executor)

	if err != nil {
		w.container.Logger.Error("Workflow execution failed:", err)
		wf.ProcessResults = types.JSONB{
			"Error": err.Error(),
		}
	}

	duration := time.Since(startTime)
	wf.StartedTime = startTime
	wf.FinishedTime = startTime.Add(duration)
	wf.CurrentAttempt = currentAttempt
	wf.Finished = cmp.Or(wf.CurrentAttempt >= w.props.MaxAttempt, w.isFinished(wf))

	w.repo.baseRepo.Update(ctx, wf)

	return true
}
