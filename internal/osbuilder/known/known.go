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

// Default project manifest file name.
const ProjectFileName = "PROJECT"

// Centralized value lists (exported) and fast lookup sets (unexported)
var (
	// AllWebFrameworks lists supported web frameworks.
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
	// AllDeploymentModes lists supported deployment modes.
	AllDeploymentModes = []string{
		DeploymentModeSystemd,
		DeploymentModeDocker,
		DeploymentModeKubernetes,
	}
	// AllApplicationTypes lists supported application types.
	AllApplicationTypes = []string{
		ApplicationTypeWebServer,
		ApplicationTypeJob,
		ApplicationTypeCLI,
	}
	// AllStorageTypes lists supported storage backends.
	AllStorageTypes = []string{
		StorageTypeMemory,
		StorageTypeMariaDB,
		StorageTypeRedis,
		StorageTypeSQLite,
		StorageTypePostgreSQL,
		StorageTypeMongo,
		StorageTypeEtcd,
	}
	// AllAppStyles lists supported scaffold styles.
	AllAppStyles = []string{
		AppStyleOneX,
		AppStyleKubernetes,
	}

	webFrameworkSet    = newSet(AllWebFrameworks)
	deploymentModeSet  = newSet(AllDeploymentModes)
	applicationTypeSet = newSet(AllApplicationTypes)
	storageTypeSet     = newSet(AllStorageTypes)
	appStyleSet        = newSet(AllAppStyles)
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
