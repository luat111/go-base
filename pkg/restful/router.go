package restful

import (
	"net/http"
	"slices"

	"github.com/gorilla/mux"
)

type Router struct {
	mux.Router
	RegisteredRoutes *[]string
}

type Middleware func(handler http.Handler) http.Handler

func NewRouter() *Router {
	muxRouter := mux.NewRouter().StrictSlash(false)
	routes := make([]string, 0)
	r := &Router{
		Router:           *muxRouter,
		RegisteredRoutes: &routes,
	}

	r.Router = *muxRouter

	return r
}

func (rou *Router) Add(method, pattern string, handler http.Handler) {
	rou.Router.NewRoute().Methods(method).Path(pattern).Handler(handler)
}

func (rou *Router) UseMiddleware(mws ...Middleware) {
	middlewares := make([]mux.MiddlewareFunc, 0, len(mws))
	for _, m := range mws {
		middlewares = append(middlewares, mux.MiddlewareFunc(m))
	}

	rou.Use(middlewares...)
}

func (rou *Router) GetRouteMethod() []string {
	var registeredMethods []string

	rou.Walk(func(route *mux.Route, _ *mux.Router, _ []*mux.Route) error {
		met, _ := route.GetMethods()
		for _, method := range met {
			if !slices.Contains(registeredMethods, method) {
				registeredMethods = append(registeredMethods, method)
			}
		}

		return nil
	})

	return registeredMethods
}
