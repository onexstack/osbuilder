package types

type MQServer struct {
	// BinaryName is the CLI binary name (e.g., "mb-apiserver").
	BinaryName string `yaml:"binaryName"`
	// WebFramework selects the framework (e.g., gin, grpc).
	WebFramework string `yaml:"webFramework"`
	// GRPCServiceName is the gRPC service name; default: UpperFirst(component name).
	GRPCServiceName string `yaml:"grpcServiceName,omitempty"`
	// StorageType selects backing storage (e.g., memory, mysql).
	StorageType string `yaml:"storageType"`
	// Feature flags
	WithHealthz     bool     `yaml:"withHealthz,omitempty"`
	WithUser        bool     `yaml:"withUser,omitempty"`
	WithOTel        bool     `yaml:"withOTel,omitempty"`
	WithWS          bool     `yaml:"withWS,omitempty"`
	WithPreloader   bool     `yaml:"withPreloader,omitempty"`
	ServiceRegistry string   `yaml:"serviceRegistry,omitempty"`
	Clients         []string `yaml:"clients,omitempty"`
	// Computed/derived fields (not serialized).
	Proj              *Project `yaml:"-"`
	Name              string   `yaml:"-"`
	EnvironmentPrefix string   `yaml:"-"`
	// APIImportPath is like: v1 "module/pkg/api/apiserver/v1"
	APIImportPath   string `yaml:"-"`
	R               *REST  `yaml:"-"`
	TypedClientName string `yaml:"-"`
}
