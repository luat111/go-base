package app

import (
	"go-base/pkg/common"
	"go-base/pkg/restful"
	"net/http"

	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
)

func (a *App[EnvInterface]) GET(pattern string, handler restful.HandlerFn) {
	a.add("GET", pattern, handler, nil)
}

func (a *App[EnvInterface]) DELETE(pattern string, handler restful.HandlerFn) {
	a.add("DELETE", pattern, handler, nil)
}

func (a *App[EnvInterface]) PUT(pattern string, validate any, handler restful.HandlerFn) {
	a.add("PUT", pattern, handler, validate)
}

func (a *App[EnvInterface]) POST(pattern string, validate any, handler restful.HandlerFn) {
	a.add("POST", pattern, handler, validate)
}

func (a *App[EnvInterface]) PATCH(pattern string, validate any, handler restful.HandlerFn) {
	a.add("PATCH", pattern, handler, validate)
}

func (a *App[EnvInterface]) add(method, pattern string, h restful.HandlerFn, validate any) {
	a.httpRegistered = true
	a.httpServer.Router.Add(method, pattern, restful.Handler{
		Function:       h,
		Container:      a.container,
		RequestTimeout: common.DefaultTimeOut,
		ValidatedBody:  validate,
	})
}

func (a *App[EnvInterface]) addMuxHandler(mux *http.ServeMux) {
	a.httpServer.Server.Handler = h2c.NewHandler(mux, &http2.Server{})
}
