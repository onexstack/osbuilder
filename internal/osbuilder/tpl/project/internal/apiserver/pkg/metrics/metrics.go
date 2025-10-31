package metrics

import (
	"context"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/metric"
)

type Metrics struct {
	Meter                     metric.Meter
	RESTResourceCreateCounter metric.Int64Counter
	RESTResourceGetCounter   metric.Int64Counter
}

var M *Metrics

// Initialize initializes Prometheus exporter and custom business metrics
func Initialize(ctx context.Context, scope string) error {
	meter := otel.Meter(scope + ".metrics")

	// Define custom metrics
	createCounter, _ := meter.Int64Counter("rest_resource_create_total", metric.WithDescription("Total number of REST resource create requests"))
	getCount, _ := meter.Int64Counter("rest_resource_get_total", metric.WithDescription("Total number of REST resource get requests"))

	// Assign global instance
	M = &Metrics{
		Meter:                     meter,
		RESTResourceCreateCounter: createCounter,
		RESTResourceGetCounter:   getCount,
	}

	return nil
}
