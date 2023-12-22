package http

import (
	"context"
	"io"
	"net/http"
	"testing"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"

	"github.com/go-chocolate/server/traceutil"
)

func newServer(t *testing.T) {
	c := Config{}
	c.Addr = ":8080"
	c.Tracing = true
	c.Logger.Enable = true
	server := New(c)
	server.Router(func(mux *http.ServeMux) {
		mux.HandleFunc("/hello", func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte("hello," + traceutil.TraceIDFromContext(r.Context())))
		})
	})
	if err := server.Run(context.Background()); err != nil {
		t.Error(err)
	}
}
func TestServer(t *testing.T) {
	otel.SetTracerProvider(sdktrace.NewTracerProvider())
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(propagation.TraceContext{}, propagation.Baggage{}))
	go newServer(t)

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
