package tracing

import (
	"context"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
)

type Provider struct {
	serviceName string
	exporterURL string
	provider    *trace.TracerProvider
}

func NewProvider(ctx context.Context, serviceName, exporterURL string) (*Provider, error) {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	// setup exporter
	e, err := otlptrace.New(ctx, otlptracegrpc.NewClient(
		otlptracegrpc.WithEndpoint(exporterURL),
		otlptracegrpc.WithInsecure(), // as connection is not need to be secured in this project
	))
	if err != nil {
		return nil, err
	}

	// serup resource
	r := resource.NewWithAttributes(
		semconv.SchemaURL,
		semconv.ServiceNameKey.String(serviceName),
		semconv.ServiceVersionKey.String("0.0.1"),
	)

	// setup sampler
	s := trace.AlwaysSample()

	tracerProvider := trace.NewTracerProvider(
		trace.WithSampler(s),
		trace.WithBatcher(e),
		trace.WithResource(r),
	)

	return &Provider{
		serviceName: serviceName,
		exporterURL: exporterURL,
		provider:    tracerProvider,
	}, nil
}

func (p *Provider) RegisterAsGlobal() (func(ctx context.Context) error, error) {
	// set global provider
	otel.SetTracerProvider(p.provider)

	// set global propagator to tracecontext (the default is no-op).
	otel.SetTextMapPropagator(propagation.TraceContext{})

	return p.provider.Shutdown, nil
}
