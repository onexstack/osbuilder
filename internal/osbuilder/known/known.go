package known

import "strings"

// Web framework identifiers used across the project (YAML/config values).
const (
	// Gin HTTP framework.
	WebFrameworkGin = "gin"
	// gRPC server framework.
	WebFrameworkGRPC = "grpc"
	// gRPC-Gateway (HTTP ↔ gRPC proxy).
	WebFrameworkGRPCGateway = "grpc-gateway"
	// Kratos microservice framework.
	WebFrameworkKratos = "kratos"
	// Go-Zero microservice framework.
	WebFrameworkGoZero = "go-zero"
	// Kitex RPC framework.
	WebFrameworkKitex = "kitex"
	// Hertz high-performance HTTP framework.
	WebFrameworkHertz = "hertz"
	// OneX framework (project-specific).
	WebFrameworkOneX = "onex"
)

// Deployment modes for releases and runtime.
const (
	DeploymentModeSystemd    = "systemd"
	DeploymentModeDocker     = "docker"
	DeploymentModeKubernetes = "kubernetes"
)

// Supported storage backends.
const (
	StorageTypeMemory     = "memory"
	StorageTypeMariaDB    = "mariadb"
	StorageTypeRedis      = "redis"
	StorageTypeSQLite     = "sqlite"
	StorageTypePostgreSQL = "postgresql"
	StorageTypeMongo      = "mongo"
	StorageTypeEtcd       = "etcd"
)

// Scaffold/application style presets.
const (
	AppStyleOneX       = "onex"
	AppStyleKubernetes = "kubernetes"
)

// Application component types.
const (
	// Web server (HTTP/gRPC) application.
	ApplicationTypeWebServer = "webserver"
	// Background job / watcher.
	ApplicationTypeJob = "job"
	// Command-line interface application.
	ApplicationTypeCLI = "cli"
)

// Makefile generation modes (enum-like string constants).
const (
	// MakefileModeNone means no Makefile will be generated or used.
	MakefileModeNone = "none"

	// MakefileModeUnstructured means using an "unstructured" (simple) Makefile layout,
	// usually a single file or a few targets for quick start scenarios.
	// Note: the identifier is "Unstructed" while the value is "unstructured".
	MakefileModeUnstructured = "unstructured"

	// MakefileModeStructured means using a "structured" Makefile layout,
	// typically organized by modules/targets for better extensibility and reuse.
	MakefileModeStructured = "structured"
)

// Define dockerfile generation modes.
const (
	// DockerfileModeNone means do not generate a Dockerfile. Users must provide
	// build/docker/<component_name>/Dockerfile manually
	DockerfileModeNone = "none"
	// DockerfileModeRuntimeOnly means generate a runtime-only Dockerfile (expects an external build artifact).
	// Suitable for local debugging or when CI/CD produces binaries separately.
	DockerfileModeRuntimeOnly = "runtime-only"
	// DockerfileModeMultiStage means generate a builder + runtime multi-stage Dockerfile.
	DockerfileModeMultiStage = "multi-stage"
	// DockerfileModeCombined means generate both variants:
	// Multi-stage: saved as "Dockerfile"
	// Runtime-only: saved as "Dockerfile.runtime-only"
	DockerfileModeCombined = "combined"
)

// Supported service registry types (used in project configuration or scaffolding).
const (
	// No service registry integration.
	ServiceRegistryNone = "none"
	// Polaris (Tencent Cloud Service Discovery/Registry).
	ServiceRegistryPolaris = "polaris"
	// Eureka service registry (Netflix OSS).
	ServiceRegistryEureka = "eureka"
	// Consul registry (HashiCorp).
	ServiceRegistryConsul = "consul"
	// Nacos registry (Alibaba Cloud).
	ServiceRegistryNacos = "nacos"
)

// Default project manifest file name.
const ProjectFileName = "PROJECT"

// Centralized value lists (exported) and fast lookup sets (unexported)
var (
	// AllWebFrameworks lists all supported web frameworks.
	AllWebFrameworks = []string{
		WebFrameworkGin,
		WebFrameworkGRPC,
		WebFrameworkGRPCGateway,
		WebFrameworkKratos,
		WebFrameworkGoZero,
		WebFrameworkKitex,
		WebFrameworkHertz,
		WebFrameworkOneX,
	}
	// AllDeploymentModes lists all supported deployment modes.
	AllDeploymentModes = []string{
		DeploymentModeSystemd,
		DeploymentModeDocker,
		DeploymentModeKubernetes,
	}
	// AllApplicationTypes lists all supported application types.
	AllApplicationTypes = []string{
		ApplicationTypeWebServer,
		ApplicationTypeJob,
		ApplicationTypeCLI,
	}
	// AllStorageTypes lists all supported storage backends.
	AllStorageTypes = []string{
		StorageTypeMemory,
		StorageTypeMariaDB,
		StorageTypeRedis,
		StorageTypeSQLite,
		StorageTypePostgreSQL,
		StorageTypeMongo,
		StorageTypeEtcd,
	}
	// AllAppStyles lists all supported scaffold styles.
	AllAppStyles = []string{
		AppStyleOneX,
		AppStyleKubernetes,
	}

	// AllMakefileModes lists all supported Makefile generation modes.
	// Useful for validation, CLI completions, or documentation output.
	AllMakefileModes = []string{
		MakefileModeNone,
		MakefileModeUnstructured,
		MakefileModeStructured,
	}

	// AllDockerfileModes lists all supported Dockerfile generation modes.
	AllDockerfileModes = []string{
		DockerfileModeNone,
		DockerfileModeRuntimeOnly,
		DockerfileModeMultiStage,
		DockerfileModeCombined,
	}

	// AllServiceRegistryTypes lists all supported service registries.
	AllServiceRegistryTypes = []string{
		ServiceRegistryNone,
		ServiceRegistryPolaris,
		ServiceRegistryEureka,
		ServiceRegistryConsul,
		ServiceRegistryNacos,
	}

	serviceRegistrySet = newSet(AllServiceRegistryTypes)
	webFrameworkSet    = newSet(AllWebFrameworks)
	deploymentModeSet  = newSet(AllDeploymentModes)
	applicationTypeSet = newSet(AllApplicationTypes)
	storageTypeSet     = newSet(AllStorageTypes)
	appStyleSet        = newSet(AllAppStyles)
	makefileModeSet    = newSet(AllMakefileModes)
)

func newSet(values []string) map[string]struct{} {
	m := make(map[string]struct{}, len(values))
	for _, v := range values {
		m[v] = struct{}{}
	}
	return m
}

func has(set map[string]struct{}, v string) bool {
	_, ok := set[v]
	return ok
}

// IsValidWebFramework reports whether v is a supported web framework.
func IsValidWebFramework(v string) bool { return has(webFrameworkSet, v) }

// IsValidDeploymentMode reports whether v is a supported deployment mode.
func IsValidDeploymentMode(v string) bool { return has(deploymentModeSet, v) }

// IsValidApplicationType reports whether v is a supported application type.
func IsValidApplicationType(v string) bool { return has(applicationTypeSet, v) }

// IsValidStorageType reports whether v is a supported storage backend.
func IsValidStorageType(v string) bool { return has(storageTypeSet, v) }

// IsValidAppStyle reports whether v is a supported scaffold style.
func IsValidAppStyle(v string) bool { return has(appStyleSet, v) }

// IsValidMakefileMode reports whether v is a supported makefile mode.
func IsValidMakefileMode(v string) bool { return has(makefileModeSet, v) }

// CanonicalWebFramework normalizes common inputs to a supported framework.
func CanonicalWebFramework(s string) (string, bool) {
	k := normalize(s)
	switch k {
	case "gin":
		return WebFrameworkGin, true
	case "grpc":
		return WebFrameworkGRPC, true
	case "grpc-gateway", "grpcgateway":
		return WebFrameworkGRPCGateway, true
	case "kratos":
		return WebFrameworkKratos, true
	case "go-zero", "gozero":
		return WebFrameworkGoZero, true
	case "kitex":
		return WebFrameworkKitex, true
	case "hertz", "heartz": // accept legacy misspelling
		return WebFrameworkHertz, true
	case "onex":
		return WebFrameworkOneX, true
	default:
		return "", false
	}
}

// CanonicalDeploymentMode normalizes inputs to a supported deployment mode.
func CanonicalDeploymentMode(s string) (string, bool) {
	k := normalize(s)
	switch k {
	case "systemd":
		return DeploymentModeSystemd, true
	case "docker":
		return DeploymentModeDocker, true
	case "kubernetes", "k8s":
		return DeploymentModeKubernetes, true
	default:
		return "", false
	}
}

// CanonicalApplicationType normalizes inputs to a supported application type.
func CanonicalApplicationType(s string) (string, bool) {
	k := normalize(s)
	switch k {
	case "webserver", "apiserver":
		return ApplicationTypeWebServer, true
	case "job", "watch":
		return ApplicationTypeJob, true
	case "cli":
		return ApplicationTypeCLI, true
	default:
		return "", false
	}
}

// CanonicalStorageType normalizes inputs to a supported storage backend.
func CanonicalStorageType(s string) (string, bool) {
	k := normalize(s)
	switch k {
	case "memory":
		return StorageTypeMemory, true
	case "mariadb", "maria":
		return StorageTypeMariaDB, true
	case "redis":
		return StorageTypeRedis, true
	case "sqlite", "sqlite3":
		return StorageTypeSQLite, true
	case "postgresql", "postgres", "pgsql":
		return StorageTypePostgreSQL, true
	case "mongo", "mongodb":
		return StorageTypeMongo, true
	case "etcd":
		return StorageTypeEtcd, true
	default:
		return "", false
	}
}

// IsValidServiceRegistry reports whether v is a supported service registry type.
func IsValidServiceRegistry(v string) bool {
	return has(serviceRegistrySet, v)
}

// CanonicalServiceRegistry normalizes common inputs to a supported service registry type.
func CanonicalServiceRegistry(s string) (string, bool) {
	k := normalize(s)
	switch k {
	case "none":
		return ServiceRegistryNone, true
	case "polaris", "tencent-polaris":
		return ServiceRegistryPolaris, true
	case "eureka":
		return ServiceRegistryEureka, true
	case "consul":
		return ServiceRegistryConsul, true
	case "nacos":
		return ServiceRegistryNacos, true
	default:
		return "", false
	}
}

// CanonicalAppStyle normalizes inputs to a supported scaffold style.
func CanonicalAppStyle(s string) (string, bool) {
	k := normalize(s)
	switch k {
	case "onex":
		return AppStyleOneX, true
	case "kubernetes", "k8s":
		return AppStyleKubernetes, true
	default:
		return "", false
	}
}

func normalize(s string) string {
	return strings.ToLower(strings.TrimSpace(s))
}
