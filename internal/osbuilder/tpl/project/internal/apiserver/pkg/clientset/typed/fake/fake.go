package {{.Web.TypedClientName}}

import (
	"context"
	"fmt"
	"log/slog"

	genericoptions "github.com/onexstack/onexstack/pkg/options"
	"resty.dev/v3"
)

// Interface defines the basic operations for a {{.Web.TypedClientName}} client.
type Interface interface {
	// C returns the underlying resty client instance.
	C() *resty.Client
	// R creates and returns a new resty request instance.
	R(ctx context.Context) *resty.Request
	// Healthz performs a health check against the client's configured endpoint.
	Healthz(ctx context.Context) error
}

// Client implements the Interface for HTTP operations using resty.
type Client struct {
	client *resty.Client
}

// NewForConfig creates a new Client instance configured with the provided options.
func NewForConfig(opts *genericoptions.RestyOptions) *Client {
	return &Client{
        {{- if .Web.WithOTel }}
        client: opts.WithTrace().NewClient(),
        {{- else}}
        client: opts.NewClient(),
        {{- end}}
	}
}

// C returns the underlying resty client instance.
func (c *Client) C() *resty.Client {
	return c.client
}

// R creates and returns a new resty request instance.
func (c *Client) R(ctx context.Context) *resty.Request {
	return c.client.SetContext(ctx).R()
}

// Healthz performs a health check by sending a GET request to the /healthz endpoint.
func (c *Client) Healthz(ctx context.Context) error {
	resp, err := c.R(ctx).Get("/healthz")
	if err != nil {
		slog.ErrorContext(ctx, "failed to send health check request", "error", err)
		return fmt.Errorf("failed to send health check reques: %w", err)
	}

	if resp.StatusCode() != 200 {
		slog.ErrorContext(ctx, "health check failed with non-200 status code",
			"status_code", resp.StatusCode(),
			"response_body", resp.String(),
		)
		return fmt.Errorf("health check failed with non-200 status codes")
	}

	slog.InfoContext(ctx, "health check successful", "status_code", resp.StatusCode())
	return nil
}
