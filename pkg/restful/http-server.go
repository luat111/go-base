package restful

import (
	"context"
	"fmt"
	"go-base/pkg/common"
	"go-base/pkg/common/utils"
	"go-base/pkg/config"
	"go-base/pkg/container"
	"go-base/pkg/logger"
	"go-base/pkg/restful/middlewares"
	"go-base/pkg/tracing"
	"net/http"
	"time"
)

type HttpServer struct {
	Server *http.Server
	Port   int
	Router *Router
	logger logger.ILogger
}

func NewHTTPServer(
	c *container.Container,
	cnf config.Config,
	port int,
	middlewareConfigs map[string]string,
) *HttpServer {
	correlationSvc := tracing.New()
	prefixPath := cnf.Get(config.API_PATH)
	log := logger.NewLogger(common.HTTPPrefix)

	r := NewRouter(prefixPath)

	r.Use(
		// 	middleware.WSHandlerUpgrade(c, s.ws),
		// 	middleware.Tracer,
		middlewares.CORS(middlewareConfigs, r.RegisteredRoutes),
		middlewares.SecureMiddleware,
		correlationSvc.CorrelationMiddleware,
		middlewares.Logging(log),
		middlewares.AttachParams,
		middlewares.TimeoutMiddleware(cnf),
	// 	middleware.Metrics(c.Metrics()),
	)

	return &HttpServer{
		Router: r,
		Port:   port,
		logger: log,
	}
}

func (s *HttpServer) Run(c *container.Container, config config.Config) {

	/* Developer Note:
	*	WebSocket connections do not inherently support authentication mechanisms.
	*	It is recommended to authenticate users before upgrading to a WebSocket connection.
	*	Hence, we are registering middlewares here, to ensure that authentication or other
	*	middleware logic is executed during the initial HTTP handshake request, prior to upgrading
	*	the connection to WebSocket, if any.
	 */

	s.Server = &http.Server{
		Addr:              fmt.Sprintf(":%d", s.Port),
		Handler:           s.Router.Engine,
		ReadHeaderTimeout: 10 * time.Second,
	}

	s.logger.Info("HTTP Server is running on", "port", s.Port)

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
