package known

import (
	"k8s.io/apimachinery/pkg/util/sets"
)

// Registries of currently supported options used for validation and scaffolding.
// Extend by uncommenting additional items below. Avoid mutating these at runtime.
var (
	// AvailableWebFrameworks lists supported web frameworks.
	AvailableWebFrameworks = sets.New(
		WebFrameworkGin,
		WebFrameworkGRPC,
		// WebFrameworkGRPCGateway,
		// WebFrameworkKratos,
		// WebFrameworkGoZero,
		// WebFrameworkKitex,
		// WebFrameworkHeartz,
	)

	// AvailableDeploymentModes lists supported deployment modes.
	AvailableDeploymentModes = sets.New(
		DeploymentModeNone,
		DeploymentModeSystemd,
		DeploymentModeDocker,
		DeploymentModeKubernetes,
	)

	// AvailableApplicationTypes lists supported application types.
	AvailableApplicationTypes = sets.New(
		ApplicationTypeWebServer,
		// ApplicationTypeWatch,
		// ApplicationTypeCli,
	)

	// AvailableStorageTypes lists supported storage backends.
	AvailableStorageTypes = sets.New(
		StorageTypeMemory,
		StorageTypeMariaDB,
		StorageTypeMySQL,
		StorageTypeSQLite,
		StorageTypePostgreSQL,
		// StorageTypeRedis,
		// StorageTypeMongo,
		// StorageTypeEtcd,
	)

	// AvailableMakefileModes lists supported makefile modes.
	AvailableMakefileModes = sets.New(
		MakefileModeNone,
		MakefileModeUnstructured,
		MakefileModeStructured,
	)

	// AvailableDockerfileModes lists supported dockerfile modes.
	AvailableDockerfileModes = sets.New(
		DockerfileModeNone,
		DockerfileModeRuntimeOnly,
		DockerfileModeMultiStage,
		DockerfileModeCombined,
	)

	// AvailableServiceRegistry lists supported service registry.
	AvailableServiceRegistry = sets.New(
		ServiceRegistryNone,
		ServiceRegistryPolaris,
		// ServiceRegistryEureka,
		// ServiceRegistryConsul,
		// ServiceRegistryNacos,
	)
)
