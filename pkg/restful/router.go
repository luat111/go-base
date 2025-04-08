package restful

import (
	"cmp"
	"net/http"
	"slices"

	"github.com/gin-gonic/gin"
)

type Router struct {
	*gin.RouterGroup
	Engine           *gin.Engine
	RegisteredRoutes *[]string
}

type Middleware func(handler http.Handler) http.Handler

func NewRouter(prefixPath string) *Router {
	gin.SetMode(gin.ReleaseMode)

	router := gin.New()
	baseGroup := router.Group(prefixPath)

	routes := make([]string, 0)
	r := &Router{
		RouterGroup:      baseGroup,
		RegisteredRoutes: &routes,
	}

	r.Engine = router

	return r
}

func (rou *Router) GetRouteMethod() []string {
	var registeredMethods []string

	for _, r := range rou.Engine.Routes() {
		if !slices.Contains(registeredMethods, r.Method) {
			registeredMethods = append(registeredMethods, r.Method)
		}
	}

	return registeredMethods
}

type RouteGroup struct {
	RouterGroup *gin.RouterGroup
	Pattern     string
	Method      string
	Handler     *Handler
}

func (rou *Router) Add(otp RouteGroup) {
	group := cmp.Or(otp.RouterGroup, rou.RouterGroup)

	var f func(relativePath string, handlers ...gin.HandlerFunc) gin.IRoutes

	switch otp.Method {
	case "GET":
		f = group.GET
	case "DELETE":
		f = group.DELETE
	case "PUT":
		f = group.PUT
	case "POST":
		f = group.POST
	case "PATCH":
		f = group.PATCH
	default:
	}

	f(otp.Pattern, gin.WrapH(otp.Handler))
}
