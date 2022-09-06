package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel"

	"github.com/Domoryonok/tracing_demo/tracing"

	"github.com/Domoryonok/tracing_demo/articles"
	"github.com/go-kit/log"
)

const (
	serviceName = "articles-service"
)

func main() {
	// init logger
	var logger log.Logger
	logger = log.NewLogfmtLogger(log.NewSyncWriter(os.Stderr))
	logger = log.With(logger, "ts", log.DefaultTimestampUTC)

	// init tracer
	tracingProvider, err := tracing.NewProvider(
		context.Background(), serviceName, os.Getenv("TRACING_BACKEND_URL"))
	if err != nil {
		_ = logger.Log("err", err)
		os.Exit(1)
	}

	// register tracing provider as a global provider
	stopTracingProvider, err := tracingProvider.RegisterAsGlobal()
	if err != nil {
		_ = logger.Log("err", err)
		os.Exit(1)
	}
	defer func() {
		if err := stopTracingProvider(context.TODO()); err != nil {
			_ = logger.Log("svc", "otlpProvider", "stop", "failed", "err", err)
			os.Exit(1)
		}
	}()

	tracer := otel.Tracer("")

	// init article service
	var as articles.Service
	as, err = articles.NewService(
		os.Getenv("SUGGESTIONS_SERVICE_HOST"),
		os.Getenv("DATA_SOURCE_FILENAME"),
		&http.Client{
			Transport: otelhttp.NewTransport(http.DefaultTransport),
			Timeout:   10 * time.Second,
		},
		tracer,
	)
	if err != nil {
		_ = logger.Log("err", err)
		os.Exit(1)
	}

	// init application handlers
	mux := http.NewServeMux()

	mux.Handle("/articles/v1/", tracing.TracingMiddleware(
		articles.MakeHandler(as, tracer), tracer))
	http.Handle("/", mux)

	// serve http server
	errs := make(chan error, 2)
	go func() {
		addr := fmt.Sprintf(":%s", os.Getenv("PORT"))
		_ = logger.Log("state", "running", "address", addr)
		errs <- http.ListenAndServe(addr, nil)
	}()

	// graceful shutdown
	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, syscall.SIGINT)
		errs <- fmt.Errorf("%s", <-c)
	}()

	_ = logger.Log("terminated", <-errs)
}
