package metrics

import (
	"context"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
)

type Metrics struct {
	Meter                     metric.Meter
	RESTResourceCreateCounter metric.Int64Counter
	RESTResourceGetCounter    metric.Int64Counter
}

var M *Metrics

// Initialize initializes Prometheus exporter and custom business metrics
func Initialize(ctx context.Context, scope string) error {
	meter := otel.Meter(scope + ".metrics")

	// Define custom metrics
	// Prometheus metric names usually follow this pattern: {subsystem}_{object}_{action}_{unit}
	createCounter, _ := meter.Int64Counter("{{.D.ProjectName | underscore}}_{{.Web.Name}}_resources_created_total", 
		metric.WithDescription("Total number of REST resource create requests"))
	getCount, _ := meter.Int64Counter("{{.D.ProjectName | underscore}}_{{.Web.Name}}_resources_retrieved_total", 
		metric.WithDescription("Total number of REST resource get requests"))

	// Assign global instance
	M = &Metrics{
		Meter:                     meter,
		RESTResourceCreateCounter: createCounter,
		RESTResourceGetCounter:    getCount,
	}

	return nil
}

// RecordResourceCreate records a REST resource create operation.
func (m *Metrics) RecordResourceCreate(ctx context.Context, resource string) {
    attrs := []attribute.KeyValue{ attribute.String("resource", resource) }
 
    m.RESTResourceCreateCounter.Add(ctx, 1, metric.WithAttributes(attrs...))
}

// RecordResourceGet records a REST resource get operation.
func (m *Metrics) RecordResourceGet(ctx context.Context, resource string) {
    attrs := []attribute.KeyValue{ attribute.String("resource", resource) }
 
    m.RESTResourceGetCounter.Add(ctx, 1, metric.WithAttributes(attrs...))
}
