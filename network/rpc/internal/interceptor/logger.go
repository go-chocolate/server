package interceptor

import (
	"context"
	"fmt"
	"strings"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/peer"
)

type LogEntity struct {
	Request    string
	Response   string
	ClientAddr string
	Metadata   map[string][]string
	Time       time.Time
	Duration   time.Duration
	Error      error
	Context    context.Context
	FullMethod string
	Stream     bool
}

func (e *LogEntity) Fields() map[string]any {
	var meta = "{"
	for k, vals := range e.Metadata {
		meta = meta + "\"" + k + "\":\"" + strings.Join(vals, ",") + "\","
	}
	if meta[len(meta)-1] == ',' {
		meta = meta[:len(meta)-1]
	}
	meta = meta + "}"

	m := map[string]any{
		"scheme":      "grpc-server",
		"request":     e.Request,
		"response":    e.Response,
		"client_addr": e.ClientAddr,
		"metadata":    meta,
		"duration":    e.Duration.String(),
		"full_method": e.FullMethod,
		"stream":      e.Stream,
	}
	return m
}

func UnaryLoggerInterceptor(record func(entity *LogEntity)) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {
		if skip(info.FullMethod) {
			return handler(ctx, req)
		}

		log := &LogEntity{}
		defer record(log)
		log.FullMethod = info.FullMethod
		log.Time = time.Now()
		log.Context = ctx
		if pe, ok := peer.FromContext(ctx); ok {
			log.ClientAddr = pe.Addr.String()
		}
		if md, ok := metadata.FromIncomingContext(ctx); ok {
			log.Metadata = md
		}
		if s, ok := req.(fmt.Stringer); ok {
			text := s.String()
			if len(text) > 512 {
				text = text[:256] + "..."
			}
			log.Request = text
		}

		rsp, err := handler(ctx, req)
		log.Duration = time.Since(log.Time)
		log.Error = err
		if err == nil {
			if s, ok := rsp.(fmt.Stringer); ok {
				text := s.String()
				if len(text) > 512 {
					text = text[:256] + "..."
				}
				log.Response = text
			}
		}
		return rsp, err
	}
}

func StreamLoggerInterceptor(record func(entity *LogEntity)) grpc.StreamServerInterceptor {
	return func(srv any, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		if skip(info.FullMethod) {
			return handler(srv, ss)
		}
		log := &LogEntity{}
		defer record(log)
		log.FullMethod = info.FullMethod
		log.Time = time.Now()
		log.Stream = true
		log.Context = ss.Context()
		if pe, ok := peer.FromContext(log.Context); ok {
			log.ClientAddr = pe.Addr.String()
		}
		if md, ok := metadata.FromIncomingContext(log.Context); ok {
			log.Metadata = md
		}
		log.Error = handler(srv, ss)
		return log.Error
	}
}

func skip(method string) bool {
	return strings.HasPrefix(method, "/grpc.health")
}
