package workflow

import (
	"context"
	"time"

	"github.com/avast/retry-go"
)

type RetryConfig struct {
	MaxAttempt uint
}

func (w *WorkflowExecutor) Execute(ctx context.Context) error {
	err := retry.Do(
		func() error {
			_, err := w.ExecuteFunc(w.Executor)

			return err
		},
		retry.MaxDelay(time.Duration(60*time.Second)),
		retry.DelayType(func(n uint, err error, config *retry.Config) time.Duration {
			return retry.BackOffDelay(n, err, config)
		}),
		retry.Attempts(w.retryConfig.MaxAttempt),
		retry.OnRetry(func(n uint, err error) {
			if n == w.retryConfig.MaxAttempt && err != nil {
				w.createWorkflow(ctx)
			}
		}),
	)

	return err
}
