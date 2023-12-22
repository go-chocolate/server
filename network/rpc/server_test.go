package rpc

import (
	"context"
	"testing"
	"time"

	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/health/grpc_health_v1"

	"github.com/go-chocolate/server/traceutil"
)

func newServer(t *testing.T) {
	c := Config{}
	c.Tracing = true
	c.Addr = "127.0.0.1:8081"

	server := New(c)
	server.WithUnaryServerInterceptor(func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {
		if id := traceutil.TraceIDFromContext(ctx); id != "" {
			t.Log("[SERVER]", id)
		} else {
			t.Error("[SERVER]", "no trace id")
		}
		return handler(ctx, req)
	})

	if err := server.Run(context.Background()); err != nil {
		t.Error(err)
		return
	}
}

func TestServer(t *testing.T) {
	otel.SetTracerProvider(sdktrace.NewTracerProvider())
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	go newServer(t)
	time.Sleep(time.Second)
	conn, err := grpc.Dial("127.0.0.1:8081",
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithStatsHandler(otelgrpc.NewClientHandler()),
	)
	if err != nil {
		t.Error(err)
		return
	}
	defer conn.Close()
	client := grpc_health_v1.NewHealthClient(conn)
	rsp, err := client.Check(context.Background(), &grpc_health_v1.HealthCheckRequest{})
	if err != nil {
		t.Error(err)
		return
	}
	t.Log(rsp.String())
}
