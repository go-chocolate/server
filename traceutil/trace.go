package traceutil

import (
	"context"

	"github.com/sirupsen/logrus"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.20.0"
	"go.opentelemetry.io/otel/trace"
)

func Setup(c Config, opts ...sdktrace.TracerProviderOption) {
	opts = append(opts, sdktrace.WithResource(buildResource(c)))
	
	if c.Sampler > 0 {
		// Set the sampling rate based on the parent span to 100%
		opts = append(opts, sdktrace.WithSampler(sdktrace.ParentBased(sdktrace.TraceIDRatioBased(c.Sampler))))
	} else {
		opts = append(opts, sdktrace.WithSampler(sdktrace.AlwaysSample()))
	}

	otel.SetTracerProvider(sdktrace.NewTracerProvider(opts...))
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))
	otel.SetErrorHandler(otel.ErrorHandlerFunc(func(err error) {
		logrus.Errorf("[otel]%v", err)
	}))
}

func buildResource(c Config) *resource.Resource {
	var attrs []attribute.KeyValue
	if c.ServiceName != "" {
		attrs = append(attrs, semconv.ServiceNameKey.String(c.ServiceName))
	}
	if c.ServiceVersion != "" {
		attrs = append(attrs, semconv.ServiceVersionKey.String(c.ServiceVersion))
	}
	for k, v := range c.Attributes {
		attrs = append(attrs, attribute.String(k, v))
	}
	return resource.NewWithAttributes(
		semconv.SchemaURL,
		attrs...,
	)
}
func TraceIDFromContext(ctx context.Context) string {
	span := trace.SpanFromContext(ctx)
	if span == nil {
		return ""
	}
	spanCtx := span.SpanContext()
	if spanCtx.HasTraceID() {
		return spanCtx.TraceID().String()
	}
	return ""
}
