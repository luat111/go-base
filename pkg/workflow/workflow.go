package workflow

import (
	"cmp"
	"context"
	"go-base/pkg"
	"go-base/pkg/common"
	"go-base/pkg/common/types"
	"go-base/pkg/container"
	"go-base/pkg/logger"
	"sync"
	"time"
)

type WorkflowProps struct {
	Name       string
	Payload    types.JSONB
	MaxAttempt int
	Schedule   string
}

type WorkflowExecutor[T ~struct{ Workflow }] struct {
	container *container.Container
	cron      *pkg.Cronjob
	repo      IWorkflowRepository[T]

	// Executor to run the workflow
	Executor    *Executor[T]
	ExecuteFunc ExecuteFunc[T]

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

func NewWorkflowExecutor[T ~struct{ Workflow }](
	ctn *container.Container,
	props WorkflowProps,
	repo IWorkflowRepository[T],
	execFn ExecuteFunc[T],
	retryConfig RetryConfig,
) *WorkflowExecutor[T] {
	logger := logger.NewLogger(common.WorkflowPrefix)
	wfExec := &WorkflowExecutor[T]{
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

	wfExec.Init()

	return wfExec
}

func (w *WorkflowExecutor[T]) Init() {
	w.container.AddCronJob(w.props.Schedule, w.props.Name, w.crawlWorkflow)
}

func (w *WorkflowExecutor[T]) crawlWorkflow() {
	ctx := context.Background()
	wfs, err := w.repo.GetRerunWorkflows(ctx)

	if err != nil {
		w.logger.Error("Failed to get rerun workflows", "err", err)
	}

	for _, wf := range wfs {
		go w.runWorkflow(ctx, wf)
	}
}

func (w *WorkflowExecutor[T]) runWorkflow(ctx context.Context, wf struct{ Workflow }) bool {
	if wf.ID == "" {
		return false
	}

	lock, errObtainLock := w.container.Locker.Obtain(ctx, wf.ID, WF_DEFAULT_TIMEOUT, nil)
	defer lock.Release(ctx)

	if errObtainLock != nil {
		w.logger.Error(errObtainLock)
		return false
	}

	var currentAttempt int
	startTime := time.Now()
	if wf.CurrentAttempt == wf.MaxAttempts {
		currentAttempt = wf.CurrentAttempt
	} else {
		currentAttempt = wf.CurrentAttempt + 1
	}

	if wf.isFinished() {
		return true
	}

	// Map for execute
	w.mu.Lock()
	w.ProcessResults, w.Payload = wf.ProcessResults, wf.Payload
	w.mu.Unlock()

	result, err := w.ExecuteFunc(ctx, w.Executor, w.repo)

	if err != nil || result != Completed {
		w.logger.Error("Workflow execution failed", "err", err)

		wf.ProcessResults = types.JSONB{
			"Error": err.Error(),
		}
	}

	duration := time.Since(startTime)

	wf.StartedTime = startTime
	wf.FinishedTime = startTime.Add(duration)
	wf.CurrentAttempt = currentAttempt
	wf.Finished = cmp.Or(wf.CurrentAttempt >= w.props.MaxAttempt, wf.isFinished())
	wf.Status = result

	w.repo.Update(ctx, wf.ID, &wf)

	return true
}

func (w *WorkflowExecutor[T]) SetProcessResults(processResults map[string]any) {
	w.mu.Lock()
	defer w.mu.Unlock()

	w.ProcessResults = processResults
}

func (w *WorkflowExecutor[T]) SetPayload(payload map[string]any) {
	w.mu.Lock()
	defer w.mu.Unlock()

	w.Payload = payload
}
