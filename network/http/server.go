package http

import (
	"context"
	"fmt"
	"net/http"
	"reflect"

	"github.com/gin-gonic/gin"
	"github.com/julienschmidt/httprouter"
	"github.com/sirupsen/logrus"

	"github.com/go-chocolate/server/cluster/endpoint"
	"github.com/go-chocolate/server/cluster/registry"
	"github.com/go-chocolate/server/network/http/internal/handler"
	"github.com/go-chocolate/server/network/http/internal/middleware"
)

type Server struct {
	srv      *http.Server
	handler  http.Handler
	config   Config
	registry registry.Registry
}

func New(c Config) *Server {
	return &Server{
		config: c,
		srv:    &http.Server{Addr: c.Addr},
	}
}

func (s *Server) WithRegistry(reg registry.Registry) {
	s.registry = reg
}

func (s *Server) Router(router func(mux *http.ServeMux)) {
	s.check()
	mux := http.NewServeMux()
	mux.HandleFunc(handler.HealthPath, handler.Health)
	router(mux)
	s.handler = mux
}

func (s *Server) Httprouter(router func(router *httprouter.Router)) {
	s.check()
	e := httprouter.New()
	e.HandlerFunc(http.MethodGet, handler.HealthPath, handler.Health)
	router(e)
	s.handler = e
}

func (s *Server) GINRouter(router func(router gin.IRouter)) {
	s.check()
	e := gin.New()
	e.GET(handler.HealthPath, func(c *gin.Context) { handler.Health(c.Writer, c.Request) })
	router(e)
	s.handler = e
}

func (s *Server) check() {
	if s.handler != nil {
		t := reflect.TypeOf(s.handler)
		logrus.Panic(fmt.Sprintf("http hander has been setted with %v", t))
	}
}

func (s *Server) middlewares() []middleware.Middleware {
	var middlewares = []middleware.Middleware{
		middleware.Recovery(),
	}
	if s.config.Logger.Enable {
		middlewares = append(middlewares, middleware.Logger())
	}
	if s.config.Tracing {
		middlewares = append(middlewares, middleware.TraceId(), middleware.Trace(s.config.ServiceName))
	}
	if s.config.Cors.Enable {
		middlewares = append(middlewares, middleware.CORS())
	}
	if s.config.RateLimit.RateLimit > 0 {
		middlewares = append(middlewares, middleware.RateLimit(s.config.RateLimit.RateLimit))
	}
	return middlewares
}

func (s *Server) Run(ctx context.Context) error {
	if s.config.TLS.Valid() {
		return s.srv.ListenAndServeTLS(s.config.TLS.Cert, s.config.TLS.Key)
	}
	middlewares := s.middlewares()
	for _, middle := range middlewares {
		s.handler = middle(s.handler)
	}

	s.srv.Handler = s.handler

	if s.registry != nil {
		end := endpoint.New(endpoint.HTTP, s.config.ServiceName, s.config.Addr, endpoint.WithHealthCheck(&endpoint.Health{
			Protocol: endpoint.HTTP,
			Path:     "/__health__",
		}))
		if err := s.registry.Register(context.Background(), end); err != nil {
			logrus.Errorf("http server registery on error: %v", err)
		}
	}

	return s.srv.ListenAndServe()
}

func (s *Server) Shutdown() {
	_ = s.srv.Shutdown(context.Background())
}
