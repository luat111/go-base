package workflow

import (
	"context"
	"go-base/pkg/tracing"
	"time"

	"github.com/avast/retry-go"
)

type RetryConfig struct {
	MaxAttempt uint
}

func (w *WorkflowExecutor[T]) Execute(ctx context.Context) error {
	id := w.generateWfId(ctx)
	ctx = tracing.WithCorrelationId(ctx, id)

	err := retry.Do(
		func() error {
			_, err := w.ExecuteFunc(ctx, w.Executor, w.repo)

			return err
		},
		// retry.Delay(time.Second/2),
		retry.MaxDelay(time.Duration(60*time.Second)),
		retry.DelayType(func(n uint, err error, config *retry.Config) time.Duration {
			return retry.BackOffDelay(n, err, config)
		}),
		retry.Attempts(w.retryConfig.MaxAttempt),
		retry.OnRetry(func(n uint, err error) {
			if n == w.retryConfig.MaxAttempt-1 && err != nil {
				w.logger.Warn("Saved workflow", "name", w.props.Name, "id", id)
				w.repo.CreateWorkflow(ctx,
					&T{
						Workflow: Workflow{
							ID:             id,
							Status:         New,
							CurrentAttempt: 1,
							WorkflowName:   w.props.Name,
							MaxAttempts:    w.props.MaxAttempt,
							ProcessResults: w.ProcessResults,
							Payload:        w.Payload,
						},
					},
				)
			} else {
				w.logger.Warn("Retry workflow", "name", w.props.Name, "id", id, "attempt", n)
			}
		}),
	)

	return err
}

func (w *WorkflowExecutor[T]) generateWfId(ctx context.Context) string {
	correlationId := tracing.FromContext(ctx)
	return w.props.Name + "_" + correlationId
}
