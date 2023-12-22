package rpc

import (
	"context"
	"net"
	"time"

	"github.com/sirupsen/logrus"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/stats"

	"github.com/go-chocolate/server/cluster/endpoint"
	"github.com/go-chocolate/server/network/rpc/internal/interceptor"

	"github.com/go-chocolate/server/cluster/registry"
)

type Server struct {
	config  Config
	options []grpc.ServerOption

	unaryServerInterceptors  []grpc.UnaryServerInterceptor
	streamServerInterceptors []grpc.StreamServerInterceptor
	registry                 registry.Registry
	grpcRegister             func(server *grpc.Server)

	server   *grpc.Server
	listener net.Listener
}

func (s *Server) Router(router func(server *grpc.Server)) {
	s.grpcRegister = router
}

func (s *Server) WithGRPCOption(options ...grpc.ServerOption) {
	s.options = append(s.options, options...)
}

func (s *Server) WithUnaryServerInterceptor(interceptors ...grpc.UnaryServerInterceptor) {
	s.unaryServerInterceptors = append(s.unaryServerInterceptors, interceptors...)
}

func (s *Server) WithStreamServerInterceptor(interceptors ...grpc.StreamServerInterceptor) {
	s.streamServerInterceptors = append(s.streamServerInterceptors, interceptors...)
}

func (s *Server) WithStatsHandler(handlers ...stats.Handler) {
	for _, handler := range handlers {
		s.options = append(s.options, grpc.StatsHandler(handler))
	}
}

func (s *Server) register() {
	if s.config.SkipRegistry || s.registry == nil {
		return
	}

	if s.config.ServiceName == "" {
		logrus.Panic("ServiceName is not specified, please check the config")
	}
	for {
		if err := s.registry.Register(context.Background(), &endpoint.Endpoint{
			Protocol: endpoint.GRPC,
			Name:     s.config.ServiceName,
			Address:  s.config.Addr,
			Health:   &endpoint.Health{Protocol: endpoint.GRPC},
		}); err != nil {
			logrus.Errorf("register rpc server failed:%s %s %v", s.config.ServiceName, s.config.Addr, err)
			time.Sleep(time.Second * 30)
		} else {
			return
		}
	}
}

func (s *Server) initialize() {
	var c = s.config
	s.unaryServerInterceptors = append(s.unaryServerInterceptors, interceptor.RecoveryUnaryServerInterceptor())
	s.streamServerInterceptors = append(s.streamServerInterceptors, interceptor.RecoveryStreamServerInterceptor())
	if c.Timeout != "" {
		if timeout, err := time.ParseDuration(c.Timeout); err == nil {
			s.options = append(s.options, grpc.ConnectionTimeout(timeout))
		}
	}
	if c.MaxRecvMsgSize != "" {
		if val := c.MaxRecvMsgSize.Value(); val > 0 {
			s.options = append(s.options, grpc.MaxRecvMsgSize(val))
		}
	}
	if c.MaxSendMsgSize != "" {
		if val := c.MaxSendMsgSize.Value(); val > 0 {
			s.options = append(s.options, grpc.MaxSendMsgSize(val))
		}
	}
	if c.Tracing {
		s.options = append(s.options, grpc.StatsHandler(otelgrpc.NewServerHandler()))
	}

	// 日志记录
	if c.Logger.Enable {
		var logger = func(entity *interceptor.LogEntity) {
			log := logrus.WithContext(entity.Context).WithFields(entity.Fields())
			if entity.Error != nil {
				log.Error(entity.Error)
			} else {
				log.Info()
			}
		}
		s.unaryServerInterceptors = append(s.unaryServerInterceptors, interceptor.UnaryLoggerInterceptor(logger))
		s.streamServerInterceptors = append(s.streamServerInterceptors, interceptor.StreamLoggerInterceptor(logger))
	}

	s.options = append(s.options,
		grpc.ChainUnaryInterceptor(s.unaryServerInterceptors...),
		grpc.ChainStreamInterceptor(s.streamServerInterceptors...),
	)
}

func (s *Server) Run(ctx context.Context) error {
	s.initialize()

	s.server = grpc.NewServer(s.options...)

	grpc_health_v1.RegisterHealthServer(s.server, health.NewServer())

	if s.grpcRegister != nil {
		s.grpcRegister(s.server)
	}

	listener, err := net.Listen("tcp", s.config.Addr)
	if err != nil {
		return err
	}
	s.listener = listener
	go s.register()
	return s.server.Serve(listener)
}

func (s *Server) Shutdown() {
	s.server.GracefulStop()
	s.listener.Close()
}

func New(c Config) *Server {
	srv := &Server{config: c}
	return srv
}
