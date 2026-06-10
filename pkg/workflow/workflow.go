package workflow

import (
	"cmp"
	"context"
	"errors"
	"go-base/pkg"
	"go-base/pkg/common"
	"go-base/pkg/common/types"
	"go-base/pkg/container"
	"go-base/pkg/logger"
	"go-base/pkg/tracing"
	"sync"
	"time"
)

type (
	WorkflowProps struct {
		Name       string
		Payload    types.JSONB
		MaxAttempt int
		Schedule   string
	}

	WorkflowExecutor[T ~struct{ Workflow }] struct {
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
)

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

	if w.container.Locker == nil {
		w.logger.Error("Locker is nil")
		return false
	}

	lockMutex, errObtainLock := w.container.Redsync.TryLock(ctx, wf.ID, WF_DEFAULT_TIMEOUT)
	defer w.container.Redsync.Unlock(lockMutex)

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
		w.logger.Error("Workflow execution failed", "id", wf.ID, "err", err)
	}

	duration := time.Since(startTime)

	wf.Duration = duration
	wf.Status = result
	wf.StartedTime = startTime
	wf.FinishedTime = startTime.Add(duration)
	wf.CurrentAttempt = currentAttempt
	wf.ProcessResults = w.ProcessResults
	wf.Finished = cmp.Or(wf.CurrentAttempt >= w.props.MaxAttempt, wf.isFinished())

	w.repo.Update(ctx, wf.ID, wf)

	return true
}

func (w *WorkflowExecutor[T]) SetProcessResults(processResults types.JSONB) {
	w.mu.Lock()
	defer w.mu.Unlock()

	w.ProcessResults = processResults
}

func (w *WorkflowExecutor[T]) SetPayload(payload map[string]any) {
	w.mu.Lock()
	defer w.mu.Unlock()

	w.Payload = payload
}

func (w *WorkflowExecutor[T]) SaveResult(ctx context.Context) error {
	id := tracing.FromContext(ctx)
	if id == "" {
		return errors.New("not found workflow id")
	}

	w.repo.Update(ctx, id, T{
		Workflow: Workflow{
			ProcessResults: w.ProcessResults,
		},
	})

	return nil
}
