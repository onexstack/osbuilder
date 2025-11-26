package {{.Web.TypedClientName}}

import (
	"fmt"

	genericoptions "github.com/onexstack/onexstack/pkg/options"
	"resty.dev/v3"
)

// Interface defines the basic operations for the client
type Interface interface {
	C() *resty.Client
	R() *resty.Request
	Hello() error
}

// Client implements the Interface for HTTP operations
type Client struct {
	client *resty.Client
}

// NewForConfig creates a new client instance based on the provided configuration
func NewForConfig(opts *genericoptions.RestyOptions) *Client {
	return &Client{
		{{- if .Web.WithOTel }}
		client: opts.WithTrace().NewClient(),
		{{- else}}
		client: opts.NewClient(),
		{{- end}}
	}
}

// C returns the underlying resty client
func (c *Client) C() *resty.Client {
	return c.client
}

// R creates and returns a new request instance
func (c *Client) R() *resty.Request {
	return c.client.R()
}

// Hello performs a health check by requesting the /healthz endpoint
func (c *Client) Hello() error {
	resp, err := c.client.R().Get("/healthz")
	if err != nil {
		return fmt.Errorf("failed to request /healthz endpoint: %w", err)
	}

	if resp.StatusCode() != 200 {
		return fmt.Errorf("health check failed with status code: %d, response: %s",
			resp.StatusCode(), resp.String())
	}

	fmt.Printf("Health check successful: %s\n", resp.String())
	return nil
}
