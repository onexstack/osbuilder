// nolint: err113
package options

import (
	{{- if or (eq .Web.WebFramework "grpc") (eq .Web.WebFramework "grpc-gateway")}}
	{{- if .Web.WithOTel}}
	"fmt"
    {{- end}}
    {{- end}}

	{{- if .Web.WithUser}}
    "time"
    "errors"
    {{- end}}
	{{- if eq .Web.ServiceRegistry "polaris" }}
    "net"
    "strconv"
    {{- end}}
	genericoptions "github.com/onexstack/onexstack/pkg/options"
	"github.com/spf13/pflag"
	utilerrors "k8s.io/apimachinery/pkg/util/errors"

	"{{.D.ModuleName}}/internal/{{.Web.Name}}"
)

// ServerOptions contains the configuration options for the server.
type ServerOptions struct {
	{{- if .Web.WithUser}}
    // JWTKey 定义 JWT 密钥.
    JWTKey string `json:"jwt-key" mapstructure:"jwt-key"`
    // Expiration 定义 JWT Token 的过期时间.
    Expiration time.Duration `json:"expiration" mapstructure:"expiration"`
    {{- end}}
	// TLSOptions contains the TLS configuration options.
	TLSOptions *genericoptions.TLSOptions `json:"tls" mapstructure:"tls"`
	{{- if or (eq .Web.WebFramework "gin") (eq .Web.WebFramework "grpc-gateway")}}
	// HTTPOptions contains the HTTP configuration options.
	HTTPOptions *genericoptions.HTTPOptions `json:"http" mapstructure:"http"`
	{{- end}}
	{{- if or (eq .Web.WebFramework "grpc") (eq .Web.WebFramework "grpc-gateway")}}
	// GRPCOptions contains the gRPC configuration options.
	GRPCOptions *genericoptions.GRPCOptions `json:"grpc" mapstructure:"grpc"`
	{{- end}}
	{{- if eq .Web.StorageType "mariadb"}}
	// MySQLOptions contains the MySQL configuration options.
	MySQLOptions *genericoptions.MySQLOptions `json:"mysql" mapstructure:"mysql"`
	{{- end}}
	{{- if eq .Web.StorageType "postgresql"}}
	// PostgreSQLOptions contains the PostgreSQL configuration options.
	PostgreSQLOptions *genericoptions.PostgreSQLOptions `json:"postgresql" mapstructure:"postgresql"`
	{{- end}}
	{{- if eq .Web.StorageType "sqlite"}}
	// SQLiteOptions contains the SQLite configuration options.
	SQLiteOptions *genericoptions.SQLiteOptions `json:"sqlite" mapstructure:"sqlite"`
	{{- end}}
	{{- if eq .Web.ServiceRegistry "polaris" }}
	// PolarisOptions used to specify the polaris options.
    PolarisOptions *genericoptions.PolarisOptions `json:"polaris" mapstructure:"polaris"`
	{{- end}}
	{{- if .Web.WithOTel}}
    // OTelOptions used to specify the otel options.
    OTelOptions *genericoptions.OTelOptions `json:"otel" mapstructure:"otel"`
	{{- if or (eq .Web.WebFramework "grpc") (eq .Web.WebFramework "grpc-gateway")}}
    // MetricsAddr specifies the address for Prometheus metrics endpoint.
    MetricsAddr string `json:"metrics-addr" mapstructure:"metrics-addr"`
	{{- end}}
	{{- else}}
    // SlogOptions used to specify the slog options.
    SlogOptions *genericoptions.SlogOptions `json:"slog" mapstructure:"slog"`
	{{- end}}
}

// NewServerOptions creates a ServerOptions instance with default values.
func NewServerOptions() *ServerOptions {
	opts := &ServerOptions{
	    {{- if .Web.WithUser}}
        JWTKey:            "",
        Expiration:        2 * time.Hour,
		{{- end}}
		TLSOptions:        genericoptions.NewTLSOptions(),
		{{- if or (eq .Web.WebFramework "gin") (eq .Web.WebFramework "grpc-gateway")}}
		HTTPOptions:       genericoptions.NewHTTPOptions(),
		{{end}}
		{{- if or (eq .Web.WebFramework "grpc") (eq .Web.WebFramework "grpc-gateway")}}
		GRPCOptions:       genericoptions.NewGRPCOptions(),
		{{- end}}
		{{- if eq .Web.StorageType "mariadb"}}
		MySQLOptions:      genericoptions.NewMySQLOptions(),
		{{- end}}
		{{- if eq .Web.StorageType "postgresql"}}
		PostgreSQLOptions:      genericoptions.NewPostgreSQLOptions(),
		{{- end}}
		{{- if eq .Web.StorageType "sqlite"}}
		SQLiteOptions:      genericoptions.NewSQLiteOptions(),
		{{- end}}
	    {{- if eq .Web.ServiceRegistry "polaris" }}
		PolarisOptions: genericoptions.NewPolarisOptions(),
		{{- end}}
		{{- if .Web.WithOTel}}
		OTelOptions: genericoptions.NewOTelOptions(),
		{{- if or (eq .Web.WebFramework "grpc") (eq .Web.WebFramework "grpc-gateway")}}
    	MetricsAddr: "0.0.0.0:29090",
		{{- end}}
		{{- else}}
		SlogOptions: genericoptions.NewSlogOptions(),
		{{- end}}
	}
	{{- if or (eq .Web.WebFramework "gin") (eq .Web.WebFramework "grpc-gateway")}}
	opts.HTTPOptions.Addr = ":5555"
	{{- end}}
	{{- if or (eq .Web.WebFramework "grpc") (eq .Web.WebFramework "grpc-gateway")}}
	opts.GRPCOptions.Addr = ":6666"
	{{- end}}

	{{- if eq .Web.ServiceRegistry "polaris" }}
	// If enable polaris register
	{{- if or (eq .Web.WebFramework "gin") (eq .Web.WebFramework "grpc-gateway")}}
	_, port, _:= net.SplitHostPort(opts.HTTPOptions.Addr)
	{{- end}}
	{{- if or (eq .Web.WebFramework "grpc") (eq .Web.WebFramework "grpc-gateway")}}
	_, port, _:= net.SplitHostPort(opts.GRPCOptions.Addr)
	{{- end}}
	intPort, _ := strconv.Atoi(port)
	opts.PolarisOptions.Provider.Port = intPort
	{{- end}}

	return opts
}

// AddFlags binds the options in ServerOptions to command-line flags.
func (o *ServerOptions) AddFlags(fs *pflag.FlagSet) {
	{{- if .Web.WithUser}}
    fs.StringVar(&o.JWTKey, "jwt-key", o.JWTKey, "JWT signing key. Must be at least 6 characters long.")
    // 绑定 JWT Token 的过期时间选项到命令行标志。
    // 参数名称为 `--expiration`，默认值为 o.Expiration
    fs.DurationVar(&o.Expiration, "expiration", o.Expiration, "The expiration duration of JWT tokens.")
	{{- end}}
	// Add command-line flags for sub-options.
	o.TLSOptions.AddFlags(fs)
	{{- if or (eq .Web.WebFramework "gin") (eq .Web.WebFramework "grpc-gateway")}}
	o.HTTPOptions.AddFlags(fs)
	{{- end}}
	{{- if or (eq .Web.WebFramework "grpc") (eq .Web.WebFramework "grpc-gateway")}}
	o.GRPCOptions.AddFlags(fs)
	{{- end}}
	{{- if eq .Web.StorageType "mariadb"}}
	o.MySQLOptions.AddFlags(fs)
	{{- end}}
	{{- if eq .Web.StorageType "postgresql"}}
	o.PostgreSQLOptions.AddFlags(fs)
	{{- end}}
	{{- if eq .Web.StorageType "sqlite"}}
	o.SQLiteOptions.AddFlags(fs)
	{{- end}}
	{{- if eq .Web.ServiceRegistry "polaris" }}
	o.PolarisOptions.AddFlags(fs)
	{{- end}}
    {{- if .Web.WithOTel}}                                                     
	o.OTelOptions.AddFlags(fs)
	{{- if or (eq .Web.WebFramework "grpc") (eq .Web.WebFramework "grpc-gateway")}}
	fs.StringVar(&o.MetricsAddr, "metrics-addr", o.MetricsAddr, "The address to expose the Prometheus /metrics endpoint.")
    {{- end}}  
    {{- else}}                                                                
	o.SlogOptions.AddFlags(fs)
    {{- end}}  
}

// Complete completes all the required options.
func (o *ServerOptions) Complete() error {
	// TODO: Add the completion logic if needed.
    return nil
}

// Validate checks whether the options in ServerOptions are valid.
func (o *ServerOptions) Validate() error {
	errs := []error{}

	{{- if .Web.WithUser}}
    // 校验 JWTKey 长度
    if len(o.JWTKey) < 6 {
        errs = append(errs, errors.New("JWTKey must be at least 6 characters long"))
    }
	{{- end}}

	// Validate sub-options.
	errs = append(errs, o.TLSOptions.Validate()...)
	{{- if or (eq .Web.WebFramework "gin") (eq .Web.WebFramework "grpc-gateway")}}
	errs = append(errs, o.HTTPOptions.Validate()...)
	{{- end}}
	{{- if or (eq .Web.WebFramework "grpc") (eq .Web.WebFramework "grpc-gateway")}}
	errs = append(errs, o.GRPCOptions.Validate()...)
	{{- end}}
	{{- if eq .Web.StorageType "mariadb"}}
	errs = append(errs, o.MySQLOptions.Validate()...)
	{{- end}}
	{{- if eq .Web.StorageType "postgresql"}}
	errs = append(errs, o.PostgreSQLOptions.Validate()...)
	{{- end}}
	{{- if eq .Web.StorageType "sqlite"}}
	errs = append(errs, o.SQLiteOptions.Validate()...)
	{{- end}}
	{{- if eq .Web.ServiceRegistry "polaris" }}
	errs = append(errs, o.PolarisOptions.Validate()...)
	{{- end}}
	{{- if .Web.WithOTel}}                                                     
	errs = append(errs, o.OTelOptions.Validate()...)
	{{- if or (eq .Web.WebFramework "grpc") (eq .Web.WebFramework "grpc-gateway")}}
    // Validate metrics address format.
    if o.MetricsAddr == "" {
        errs = append(errs, fmt.Errorf("metrics-addr cannot be empty"))
    }
    {{- end}}
    {{- else}}                                                                

	errs = append(errs, o.SlogOptions.Validate()...)
    {{- end}}

	// Aggregate all errors and return them.
	return utilerrors.NewAggregate(errs)
}

// Config builds an {{.Web.Name}}.Config based on ServerOptions.
func (o *ServerOptions) Config() (*{{.Web.Name}}.Config, error) {
	return &{{.Web.Name}}.Config{
	    {{- if .Web.WithUser}}
        JWTKey:            o.JWTKey,
        Expiration:        o.Expiration,
	    {{- end}}
		TLSOptions:        o.TLSOptions,
		{{- if or (eq .Web.WebFramework "gin") (eq .Web.WebFramework "grpc-gateway")}}
		HTTPOptions:       o.HTTPOptions,
		{{- end}}
		{{- if or (eq .Web.WebFramework "grpc") (eq .Web.WebFramework "grpc-gateway")}}
		GRPCOptions:       o.GRPCOptions,
		{{- end}}
		{{- if eq .Web.StorageType "mariadb"}}
		MySQLOptions:      o.MySQLOptions,
		{{- end}}
		{{- if eq .Web.StorageType "postgresql"}}
		PostgreSQLOptions:      o.PostgreSQLOptions,
		{{- end}}
		{{- if eq .Web.StorageType "sqlite"}}
		SQLiteOptions:      o.SQLiteOptions,
		{{- end}}
		{{- if eq .Web.ServiceRegistry "polaris" }}
	    PolarisOptions: o.PolarisOptions,
		{{- end}}
		{{- if or (eq .Web.WebFramework "grpc") (eq .Web.WebFramework "grpc-gateway")}}
		{{- if .Web.WithOTel}}                                                     
		MetricsAddr: o.MetricsAddr,
		{{- end}}
		{{- end}}
	}, nil
}
