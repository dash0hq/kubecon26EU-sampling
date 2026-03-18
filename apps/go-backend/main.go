package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"math/rand/v2"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
	"go.opentelemetry.io/otel/trace"
)

var logger = slog.New(slog.NewJSONHandler(os.Stdout, nil))

func initTelemetry(ctx context.Context) (func(), error) {
	res, err := resource.New(ctx,
		resource.WithAttributes(semconv.ServiceNameKey.String("go-backend")),
		resource.WithFromEnv(),
	)
	if err != nil {
		return nil, fmt.Errorf("resource: %w", err)
	}

	traceExp, err := otlptracegrpc.New(ctx)
	if err != nil {
		return nil, fmt.Errorf("trace exporter: %w", err)
	}
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(traceExp),
		sdktrace.WithResource(res),
	)
	otel.SetTracerProvider(tp)

	return func() {
		_ = tp.Shutdown(context.Background())
	}, nil
}

func rollHandler(w http.ResponseWriter, r *http.Request) {
	span := trace.SpanFromContext(r.Context())

	value := rand.IntN(6) + 1
	span.SetAttributes(attribute.Int("dice.value", value))

	switch value {
	case 1:
		// Simulate an error: rolling a 1 is unlucky.
		err := fmt.Errorf("unlucky roll: value 1 is forbidden")
		span.SetStatus(codes.Error, err.Error())
		logger.ErrorContext(r.Context(), "roll.failed",
			"dice.value", value,
			"error", err.Error(),
		)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(w, `{"error":"unlucky roll"}`)
		return

	case 6:
		// Simulate a slow request: rolling a 6 triggers a 2.5 s delay.
		logger.InfoContext(r.Context(), "roll.slow", "dice.value", value, "delay_ms", 2500)
		time.Sleep(2500 * time.Millisecond)
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(map[string]int{"value": value}); err != nil {
		logger.ErrorContext(r.Context(), "encode.failed", "error", err.Error())
	}
}

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT)
	defer stop()

	shutdown, err := initTelemetry(ctx)
	if err != nil {
		logger.Error("telemetry init failed", "error", err.Error())
		os.Exit(1)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /roll", rollHandler)
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"status":"ok"}`)
	})

	handler := otelhttp.NewHandler(mux, "go-backend",
		otelhttp.WithSpanNameFormatter(func(_ string, r *http.Request) string {
			return r.Method + " " + r.URL.Path
		}),
	)
	srv := &http.Server{Addr: ":8080", Handler: handler}

	go func() {
		logger.Info("go-backend listening", "addr", ":8080")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("server error", "error", err.Error())
		}
	}()

	<-ctx.Done()
	logger.Info("shutting down")
	_ = srv.Shutdown(context.Background())
	shutdown()
}
