package telemetry

import (
	"context"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/sdk/trace"
)

type ShutdownFn func(context.Context) error

func InitOTel(enabled bool) ShutdownFn {
	if !enabled {
		return func(context.Context) error { return nil }
	}
	tp := trace.NewTracerProvider()
	otel.SetTracerProvider(tp)
	return tp.Shutdown
}
