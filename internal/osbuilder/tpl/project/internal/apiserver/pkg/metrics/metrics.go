package metrics

import (
	"context"
	"sync"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
)

// Metrics holds the OpenTelemetry instruments for capturing application metrics.
type Metrics struct {
	Meter                     metric.Meter
	RESTResourceCreateCounter metric.Int64Counter
	RESTResourceGetCounter    metric.Int64Counter
}

var (
	// M is the global metrics instance.
	M *Metrics
	// once ensures the initialization logic runs only once.
	once sync.Once
)

// Init initializes the global metrics instance using the singleton pattern.
func Init(scope string) error {
	once.Do(func() {
		meter := otel.Meter(scope)

		// Define custom metrics.
    	createCounter, _ := meter.Int64Counter(
        	"{{.D.ProjectName | underscore}}_{{.Web.Name}}_resources_created_total",
        	metric.WithDescription("Total number of REST resource create requests"),
    	)

    	getCount, _ := meter.Int64Counter(
        	"{{.D.ProjectName | underscore}}_{{.Web.Name}}_resources_retrieved_total",
        	metric.WithDescription("Total number of REST resource get requests"),
    	)

		// Assign the global singleton.
		M = &Metrics{
			Meter:                     meter,
			RESTResourceCreateCounter: createCounter,
			RESTResourceGetCounter:    getCount,
		}
	})

	return nil
}

// RecordResourceCreate increments the counter for a resource creation operation.
func (m *Metrics) RecordResourceCreate(ctx context.Context, resource string) {
	attrs := []attribute.KeyValue{attribute.String("resource", resource)}

	m.RESTResourceCreateCounter.Add(ctx, 1, metric.WithAttributes(attrs...))
}

// RecordResourceGet increments the counter for a resource retrieval operation.
func (m *Metrics) RecordResourceGet(ctx context.Context, resource string) {
	attrs := []attribute.KeyValue{attribute.String("resource", resource)}

	m.RESTResourceGetCounter.Add(ctx, 1, metric.WithAttributes(attrs...))
}
