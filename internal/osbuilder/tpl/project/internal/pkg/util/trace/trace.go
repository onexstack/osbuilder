package trace

import (
	"context"

	"github.com/twmb/franz-go/pkg/kgo"
	"go.opentelemetry.io/otel"
)

// FranzHeaderCarrier adapts kgo.RecordHeader slices to implement the
// propagation.TextMapCarrier interface, allowing OpenTelemetry context
// to be injected into and extracted from Kafka message headers.
type FranzHeaderCarrier []kgo.RecordHeader

// Get retrieves the value associated with the given key from the Kafka headers.
// It returns an empty string if the key is not found.
func (h FranzHeaderCarrier) Get(key string) string {
	for _, header := range h {
		if header.Key == key {
			return string(header.Value)
		}
	}
	return ""
}

// Set adds or updates a key-value pair in the Kafka headers.
// If the key already exists, its value is updated. Otherwise, a new header is appended.
func (h *FranzHeaderCarrier) Set(key string, value string) {
	for i := range *h {
		if (*h)[i].Key == key {
			(*h)[i].Value = []byte(value) // Update existing header
			return
		}
	}
	*h = append(*h, kgo.RecordHeader{Key: key, Value: []byte(value)}) // Append new header
}

// Keys returns a slice of all unique keys present in the Kafka headers.
func (h FranzHeaderCarrier) Keys() []string {
	keys := make([]string, 0, len(h))
	seen := make(map[string]struct{})
	for _, header := range h {
		if _, ok := seen[header.Key]; !ok {
			keys = append(keys, header.Key)
			seen[header.Key] = struct{}{}
		}
	}
	return keys
}

// Inject propagates the OpenTelemetry context from `ctx` into the provided
// `kgo.RecordHeader` slice. This is typically used on the producer side
// to inject trace context into outgoing Kafka messages.
func Inject(ctx context.Context, headers *[]kgo.RecordHeader) {
	carrier := (*FranzHeaderCarrier)(headers) // Type assertion to use the adapter
	otel.GetTextMapPropagator().Inject(ctx, carrier)
}

// Extract retrieves the OpenTelemetry context from the provided `kgo.RecordHeader` slice.
// This is typically used on the consumer side to extract trace context from
// incoming Kafka messages and reconstruct the parent context.
func Extract(ctx context.Context, headers []kgo.RecordHeader) context.Context {
	carrier := FranzHeaderCarrier(headers)                    // Type assertion to use the adapter
	return otel.GetTextMapPropagator().Extract(ctx, &carrier) // Pass address of local carrier
}
