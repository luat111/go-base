package restful

import (
	"context"
	"go-base/pkg/container"
	"go-base/pkg/logger"
	"go-base/pkg/tracing"
)

type Context struct {
	Context   context.Context
	Container *container.Container
	Request   IRequest
	Responder IResponder
}

func NewContext(w IResponder, r IRequest, c *container.Container) *Context {
	return &Context{
		Context:   r.Context(),
		Request:   r,
		Responder: w,
		Container: c,
	}
}

func (c *Context) GetTrackingId() string {
	return tracing.FromContext(c.Context)
}

func (c *Context) Logger() logger.ILogger {
	return c.Container.Logger
}
