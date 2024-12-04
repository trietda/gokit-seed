package otel

import (
	"context"
	"errors"
	"os"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploghttp"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/log/global"
	"go.opentelemetry.io/otel/propagation"
	sdklog "go.opentelemetry.io/otel/sdk/log"
	"go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/trace"
)

// setupOTelSDK bootstraps the OpenTelemetry pipeline.
// If it does not return an error, make sure to call shutdown for proper cleanup.
func SetupOTelSdk(lgp *sdklog.LoggerProvider) (func(context.Context) error, error) {
	connectCtx := context.Background()
	shutdownCtx := context.Background()

	var shutdownFuncs []func(context.Context) error

	// shutdown calls cleanup functions registered via shutdownFuncs.
	// The errors from the calls are joined.
	// Each registered cleanup will be invoked once.
	shutdown := func(ctx context.Context) error {
		var err error
		for _, fn := range shutdownFuncs {
			err = errors.Join(err, fn(ctx))
		}
		shutdownFuncs = nil
		return err
	}
	var err error

	// handleErr calls shutdown for cleanup and makes sure that all errors are returned.
	handleErr := func(inErr error, shutdownCtx context.Context) {
		err = errors.Join(inErr, shutdown(shutdownCtx))
	}

	// Set up propagator.
	prop := newPropagator()
	otel.SetTextMapPropagator(prop)

	// Set up trace provider.
	tracerProvider, err := newTraceProvider(connectCtx)
	if err != nil {
		handleErr(err, shutdownCtx)
		return shutdown, err
	}
	if tracerProvider != nil {
		shutdownFuncs = append(shutdownFuncs, tracerProvider.Shutdown)
		otel.SetTracerProvider(tracerProvider)
	}

	// Set up meter provider.
	meterProvider, err := newMeterProvider(connectCtx)
	if err != nil {
		handleErr(err, shutdownCtx)
		return shutdown, err
	}
	if meterProvider != nil {
		shutdownFuncs = append(shutdownFuncs, meterProvider.Shutdown)
		otel.SetMeterProvider(meterProvider)
	}

	// Set up logger provider.
	if lgp == nil {
		loggerProvider, err := NewLoggerProvider(connectCtx)
		if err != nil {
			handleErr(err, shutdownCtx)
			return shutdown, err
		}
		if loggerProvider != nil {
			shutdownFuncs = append(shutdownFuncs, loggerProvider.Shutdown)
			global.SetLoggerProvider(lgp)
		}
	} else {
		shutdownFuncs = append(shutdownFuncs, lgp.Shutdown)
		global.SetLoggerProvider(lgp)
	}

	return shutdown, nil
}

func newPropagator() propagation.TextMapPropagator {
	return propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	)
}

func newTraceProvider(ctx context.Context) (*trace.TracerProvider, error) {

	if traceEndpoint := os.Getenv("OTEL_EXPORTER_OTLP_TRACES_ENDPOINT"); traceEndpoint == "" {
		return nil, nil
	}

	traceExporter, err := otlptracehttp.New(ctx)

	if err != nil {
		return nil, err
	}

	traceProvider := trace.NewTracerProvider(
		trace.WithBatcher(traceExporter,
			// Default is 5s. Set to 1s for demonstrative purposes.
			trace.WithBatchTimeout(time.Second)),
	)
	return traceProvider, nil
}

func newMeterProvider(ctx context.Context) (*metric.MeterProvider, error) {

	if metricEndpoint := os.Getenv("OTEL_EXPORTER_OTLP_METRICS_ENDPOINT"); metricEndpoint == "" {
		return nil, nil
	}

	metricExporter, err := otlpmetrichttp.New(ctx)

	if err != nil {
		return nil, err
	}

	meterProvider := metric.NewMeterProvider(
		metric.WithReader(metric.NewPeriodicReader(metricExporter,
			// Default is 1m. Set to 3s for demonstrative purposes.
			metric.WithInterval(3*time.Second))),
	)
	return meterProvider, nil
}

func NewLoggerProvider(ctx context.Context) (*sdklog.LoggerProvider, error) {
	var (
		logExporter sdklog.Exporter
		err         error
	)

	if logEndpoint := os.Getenv("OTEL_EXPORTER_OTLP_LOGS_ENDPOINT"); logEndpoint == "" {
		return nil, nil
	}

	logExporter, err = otlploghttp.New(ctx)

	if err != nil {
		return nil, err
	}

	loggerProvider := sdklog.NewLoggerProvider(
		sdklog.WithProcessor(sdklog.NewBatchProcessor(logExporter)),
	)
	return loggerProvider, nil
}
