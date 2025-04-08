package app

import (
	"go-base/pkg/restful"

	"github.com/gin-gonic/gin"
)

func (a *App[EnvInterface]) GET(group *gin.RouterGroup, pattern string, handler restful.HandlerFn) {
	a.add(group, "GET", pattern, handler, nil)
}

func (a *App[EnvInterface]) DELETE(group *gin.RouterGroup, pattern string, handler restful.HandlerFn) {
	a.add(group, "DELETE", pattern, handler, nil)
}

func (a *App[EnvInterface]) PUT(group *gin.RouterGroup, pattern string, validate any, handler restful.HandlerFn) {
	a.add(group, "PUT", pattern, handler, validate)
}

func (a *App[EnvInterface]) POST(group *gin.RouterGroup, pattern string, validate any, handler restful.HandlerFn) {
	a.add(group, "POST", pattern, handler, validate)
}

func (a *App[EnvInterface]) PATCH(group *gin.RouterGroup, pattern string, validate any, handler restful.HandlerFn) {
	a.add(group, "PATCH", pattern, handler, validate)
}

func (a *App[EnvInterface]) BaseGroup() *gin.RouterGroup {
	return a.httpServer.Router.RouterGroup
}

func (a *App[EnvInterface]) Group(pattern string, handlers ...gin.HandlerFunc) *gin.RouterGroup {
	return a.httpServer.Router.Group(pattern, handlers...)
}

func (a *App[EnvInterface]) add(group *gin.RouterGroup, method, pattern string, h restful.HandlerFn, validate any) {
	a.httpRegistered = true

	hdl := &restful.Handler{
		Function:      h,
		Container:     a.container,
		ValidatedBody: validate,
	}

	r := restful.RouteGroup{
		RouterGroup: group,
		Pattern:     pattern,
		Method:      method,
		Handler:     hdl,
	}

	a.httpServer.Router.Add(r)
}
