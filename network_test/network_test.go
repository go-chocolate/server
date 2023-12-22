package network_test

import (
	"context"
	"io"
	"net/http"
	"testing"

	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/health/grpc_health_v1"

	httpx "github.com/go-chocolate/server/network/http"
	"github.com/go-chocolate/server/network/rpc"
	"github.com/go-chocolate/server/traceutil"
)

func init() {
	otel.SetTracerProvider(sdktrace.NewTracerProvider())
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))
}

func newRpcServer(t *testing.T, re *recorder) {
	c := rpc.Config{}
	c.Tracing = true
	c.Addr = "127.0.0.1:8081"

	server := rpc.New(c)
	server.WithUnaryServerInterceptor(func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {
		if id := traceutil.TraceIDFromContext(ctx); id != "" {
			t.Log("[SERVER]", id)
			if id != re.traceId {
				t.Errorf("trace id not match: %s %s", id, re.traceId)
			}
		} else {
			t.Error("[SERVER]", "no trace id")
		}
		return handler(ctx, req)
	})

	if err := server.Run(context.Background()); err != nil {
		t.Fatal(err)
		return
	}
}

func newHttpServer(t *testing.T, re *recorder) {
	c := httpx.Config{}
	c.Addr = ":8080"
	c.Tracing = true
	c.Logger.Enable = true
	server := httpx.New(c)
	server.Router(func(mux *http.ServeMux) {
		mux.HandleFunc("/hello", func(w http.ResponseWriter, r *http.Request) {
			re.traceId = traceutil.TraceIDFromContext(r.Context())
			conn, err := grpc.DialContext(r.Context(), "127.0.0.1:8081",
				grpc.WithTransportCredentials(insecure.NewCredentials()),
				grpc.WithStatsHandler(otelgrpc.NewClientHandler()))
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			defer conn.Close()
			client := grpc_health_v1.NewHealthClient(conn)
			//rsp, err := client.Check(context.Background(), &grpc_health_v1.HealthCheckRequest{})
			rsp, err := client.Check(r.Context(), &grpc_health_v1.HealthCheckRequest{})
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			w.Write([]byte(rsp.String()))
		})
	})
	if err := server.Run(context.Background()); err != nil {
		t.Fatal(err)
	}
}

type recorder struct {
	traceId string
}

func TestNetwork(t *testing.T) {
	re := &recorder{}
	go newHttpServer(t, re)
	go newRpcServer(t, re)

	response, err := otelhttp.Get(context.Background(), "http://127.0.0.1:8080/hello")
	if err != nil {
		t.Error(err)
		return
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		t.Error(response.StatusCode)
	}
	for k := range response.Header {
		t.Log(k, response.Header.Get(k))
	}
	b, _ := io.ReadAll(response.Body)
	t.Log(string(b))
}
