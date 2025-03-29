package restful

import (
	"cmp"
	"context"
	"errors"
	"fmt"
	"go-base/pkg/container"
	"go-base/pkg/logger"
	"net/http"
	"strings"
	"time"

	"github.com/go-playground/validator/v10"
)

var reqValidator = validator.New()

type HandlerFn func(c *Context) (any, error)

type Handler struct {
	Function       HandlerFn
	Container      *container.Container
	RequestTimeout time.Duration
	ValidatedBody  any
}

func (h Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	c := NewContext(NewResponder(w, r.Method), NewRequest(r), h.Container)

	done := make(chan struct{})
	panicked := make(chan struct{})

	var (
		result any
		err    error
	)

	go func() {
		defer func() {
			panicRecoveryHandler(recover(), h.Container.Logger, panicked)
		}()

		err = h.validateStruct(c)

		if err == nil {
			result, err = h.Function(c)
		}

		close(done)
	}()

	select {
	case <-c.Context.Done():
		// If the context's deadline has been exceeded, return a timeout error response
		if errors.Is(c.Context.Err(), context.DeadlineExceeded) {
			err = errors.New("Timeout")
		}
	case <-done:
	case <-panicked:
		err = errors.New("Panic")
	}

	// Handler function completed
	c.Responder.Respond(result, err)
}

func panicRecoveryHandler(re any, log logger.ILogger, panicked chan struct{}) {
	if re == nil {
		return
	}

	close(panicked)
	log.Error(fmt.Sprint(re))
}

type ValidateError struct {
	message string
}

func (vErr *ValidateError) Error() string {
	return vErr.message
}

func (vErr *ValidateError) StatusCode() int {
	return http.StatusBadRequest
}

func (h *Handler) validateStruct(c *Context) error {
	if h.ValidatedBody == nil {
		return nil
	}

	err := cmp.Or(c.Request.Bind(h.ValidatedBody), reqValidator.Struct(h.ValidatedBody))

	if err != nil {
		var ve validator.ValidationErrors
		var errField []string

		if errors.As(err, &ve) {
			for _, e := range err.(validator.ValidationErrors) {
				errStr := fmt.Sprintf("'%s' on tag '%s'", e.Field(), e.Tag())
				errField = append(errField, errStr)
			}

			return &ValidateError{message: fmt.Sprintf("Validate failed: %s", strings.Join(errField, ","))}
		}
	}

	return nil
}
