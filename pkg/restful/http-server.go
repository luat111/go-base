package restful

import (
	"context"
	"fmt"
	"go-base/pkg/common/utils"
	"go-base/pkg/container"
	"go-base/pkg/restful/middlewares"
	"go-base/pkg/tracing"
	"net/http"
	"time"

	"github.com/unrolled/secure"
)

type HttpServer struct {
	Server *http.Server
	Port   int
	Router *Router
}

func NewHTTPServer(Port int) *HttpServer {
	r := NewRouter()

	return &HttpServer{
		Router: r,
		Port:   Port,
	}
}

func (s *HttpServer) Run(c *container.Container, middlewareConfigs map[string]string) {

	/* Developer Note:
	*	WebSocket connections do not inherently support authentication mechanisms.
	*	It is recommended to authenticate users before upgrading to a WebSocket connection.
	*	Hence, we are registering middlewares here, to ensure that authentication or other
	*	middleware logic is executed during the initial HTTP handshake request, prior to upgrading
	*	the connection to WebSocket, if any.
	 */

	correlationSvc := tracing.New()
	secureMiddleware := secure.New(secure.Options{
		FrameDeny:        true,
		BrowserXssFilter: true,
	})

	s.Router.Use(
		// 	middleware.WSHandlerUpgrade(c, s.ws),
		// 	middleware.Tracer,
		middlewares.CORS(middlewareConfigs, s.Router.RegisteredRoutes),
		secureMiddleware.Handler,
		correlationSvc.CorrelationMiddleware,
		middlewares.Logging(c.Logger),
	// 	middleware.Metrics(c.Metrics()),
	)

	s.Server = &http.Server{
		Addr:              fmt.Sprintf(":%d", s.Port),
		Handler:           s.Router,
		ReadHeaderTimeout: 10 * time.Second,
	}

	c.Logger.Info("HTTP Server is running on", "port", s.Port)

	if err := s.Server.ListenAndServe(); err != nil {
		c.Logger.Debug(err)
	}
}

func (s *HttpServer) Shutdown(ctx context.Context) error {
	if s.Server == nil {
		return nil
	}

	err := utils.GracefulShutDown(ctx, func(ctx context.Context) error {
		return s.Server.Shutdown(ctx)
	})

	return err
}

func (s *HttpServer) MappingRoutes() {
	methods := s.Router.GetRouteMethod()

	*s.Router.RegisteredRoutes = methods
}
