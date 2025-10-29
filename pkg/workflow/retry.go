package workflow

import (
	"context"
	"time"

	"github.com/avast/retry-go"
)

type RetryConfig struct {
	MaxAttempt uint
}

func (w *WorkflowExecutor[T]) Execute(ctx context.Context) error {
	err := retry.Do(
		func() error {
			_, err := w.ExecuteFunc(ctx, w.Executor)

			return err
		},
		retry.MaxDelay(time.Duration(60*time.Second)),
		retry.DelayType(func(n uint, err error, config *retry.Config) time.Duration {
			return retry.BackOffDelay(n, err, config)
		}),
		retry.Attempts(w.retryConfig.MaxAttempt),
		retry.OnRetry(func(n uint, err error) {
			if n == w.retryConfig.MaxAttempt && err != nil {
				w.repo.CreateWorkflow(ctx, &T{
					Workflow: &Workflow{
						Status:         New,
						CurrentAttempt: 1,
						WorkflowName:   w.props.Name,
						MaxAttempts:    w.props.MaxAttempt,
						ProcessResults: w.ProcessResults,
						Payload:        w.Payload,
					},
				})
			}
		}),
	)

	return err
}
