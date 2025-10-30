package create

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"text/template"

	"github.com/fatih/color"
	stringsutil "github.com/onexstack/onexstack/pkg/util/strings"
	"github.com/spf13/cobra"
	"k8s.io/cli-runtime/pkg/genericiooptions"
	"k8s.io/kubectl/pkg/util/templates"

	cmdutil "github.com/onexstack/osbuilder/internal/osbuilder/cmd/util"
	"github.com/onexstack/osbuilder/internal/osbuilder/file"
	"github.com/onexstack/osbuilder/internal/osbuilder/helper"
	"github.com/onexstack/osbuilder/internal/osbuilder/known"
	"github.com/onexstack/osbuilder/internal/osbuilder/types"
	"github.com/onexstack/osbuilder/internal/osbuilder/validation"
)

// ProjectOptions contains flags and runtime state for the 'create project' command.
type ProjectOptions struct {
	RootDir string
	Config  string

	Project *types.Project

	genericiooptions.IOStreams
}

var (
	projectLongDesc = templates.LongDesc(`
Create a new project from a configuration file.

This command scaffolds a project structure, generates boilerplate files,
and prepares build and run artifacts.`)

	projectExamples = templates.Examples(`
  # Generate a project using a config file
  osbuilder create project ./my-project --config ./onexstack.yaml

  # Generate a project in the current directory with default config path
  osbuilder create project .`)
)

// NewProjectOptions returns a new ProjectOptions with default IO streams.
func NewProjectOptions(ioStreams genericiooptions.IOStreams) *ProjectOptions {
	return &ProjectOptions{IOStreams: ioStreams}
}

// NewCmdProject returns the 'create project' command definition.
func NewCmdProject(f cmdutil.Factory, ioStreams genericiooptions.IOStreams) *cobra.Command {
	opts := NewProjectOptions(ioStreams)

	cmd := &cobra.Command{
		Use:                   "project [DIR]",
		Aliases:               []string{"proj"},
		Short:                 "Create a new project from a config file",
		Long:                  projectLongDesc,
		Example:               projectExamples,
		DisableFlagsInUseLine: true,
		TraverseChildren:      true,
		SilenceUsage:          true,
		SilenceErrors:         true,
		// DIR is optional: defaults to current working directory.
		Args: cobra.MaximumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			cmdutil.CheckErr(opts.Complete(f, cmd, args))
			cmdutil.CheckErr(opts.Validate(cmd, args))
			cmdutil.CheckErr(opts.Run(f, opts.IOStreams, args))
		},
	}

	cmd.Flags().StringVarP(&opts.Config, "config", "c", opts.Config, "Path to project config file (default: ./onexstack.yaml under the chosen directory)")

	return cmd
}

// Complete resolves input arguments and loads project metadata.
func (opts *ProjectOptions) Complete(_ cmdutil.Factory, _ *cobra.Command, args []string) error {
	// Resolve root directory (explicit arg or current working directory)
	opts.RootDir, _ = os.Getwd()
	if len(args) == 1 {
		abs, err := filepath.Abs(args[0])
		if err != nil {
			return fmt.Errorf("resolve directory: %w", err)
		}
		opts.RootDir = abs
	}

	// Default config path if not provided (after RootDir is known)
	if strings.TrimSpace(opts.Config) == "" {
		opts.Config = filepath.Join(opts.RootDir, known.ProjectFileName)
	}

	// Load project configuration
	proj, err := LoadProjectFromFile(opts.Config)
	if err != nil {
		return fmt.Errorf("load project from config %q: %w", opts.Config, err)
	}
	// Fill generated data
	proj.D = (&types.GeneratedData{
		WorkDir:    opts.RootDir,
		APIVersion: "v1",
		APIAlias:   "v1",
		ModuleName: MustModulePath(proj.Metadata.ModulePath, opts.RootDir),
	}).Complete()

	proj.D.ProjectName = filepath.Base(opts.RootDir)
	proj.D.RegistryPrefix = proj.Metadata.Image.RegistryPrefix

	opts.Project = correctProjectConfig(proj)
	return nil
}

// Validate ensures options and loaded project are consistent.
func (opts *ProjectOptions) Validate(_ *cobra.Command, _ []string) error {
	if opts.Project == nil {
		return fmt.Errorf("project is not loaded")
	}

	// Validate module path format
	if err := validation.ValidateModulePath(opts.Project.D.ModuleName); err != nil {
		return fmt.Errorf("invalid module path %q: %w", opts.Project.D.ModuleName, err)
	}

	if err := validateMetadata(opts.Project.Metadata); err != nil {
		return err
	}

	// Validate application type (project-level)
	if len(opts.Project.Jobs) > 0 || len(opts.Project.CLIApps) > 0 {
		return fmt.Errorf(
			"unsupported application type, supported: %s",
			strings.Join(known.AvailableApplicationTypes.UnsortedList(), ", "),
		)
	}

	// Validate web frameworks and storage types (per web server)
	for _, ws := range opts.Project.WebServers {
		// Web framework
		wf := strings.TrimSpace(ws.WebFramework)
		if !known.AvailableWebFrameworks.Has(wf) {
			return fmt.Errorf(
				"web server %q: unsupported webFramework %q; supported: %s",
				ws.Name, wf, strings.Join(known.AvailableWebFrameworks.UnsortedList(), ", "),
			)
		}

		// Storage type
		st := strings.TrimSpace(ws.StorageType)
		if !known.AvailableStorageTypes.Has(st) {
			return fmt.Errorf(
				"web server %q: unsupported storageType %q; supported: %s",
				ws.Name, st, strings.Join(known.AvailableStorageTypes.UnsortedList(), ", "),
			)
		}

		// Service registry type
		sr := strings.TrimSpace(ws.ServiceRegistry)
		if !known.AvailableServiceRegistry.Has(sr) {
			return fmt.Errorf(
				"web server %q: unsupported serviceRegistry %q; supported: %s",
				ws.Name, st, strings.Join(known.AvailableServiceRegistry.UnsortedList(), ", "),
			)
		}
	}

	return nil
}

// Run generates the project files and prints next steps.
func (opts *ProjectOptions) Run(f cmdutil.Factory, ioStreams genericiooptions.IOStreams, _ []string) (err error) {
	defer func() { helper.RecordOSBuilderUsage("project", err) }()

	fm := file.NewFileManager(opts.RootDir, false)

	if err := opts.Generate(f, fm); err != nil {
		return err
	}

	// Copy third-party dependencies
	fs := helper.NewFileSystem("/")
	if err := fm.CopyFiles(filepath.Join(fs.BasePath, "/project/third_party"), filepath.Join(opts.RootDir, "third_party")); err != nil {
		return fmt.Errorf("copy third_party: %w", err)
	}

	opts.PrintGettingStarted()
	return nil
}

// PrintGettingStarted prints follow-up commands to bootstrap the project.
func (opts *ProjectOptions) PrintGettingStarted() {
	baseName := filepath.Base(opts.Project.D.WorkDir)

	fmt.Printf("\n🍺 Project creation succeeded %s\n", color.GreenString(baseName))
	fmt.Print("💻 Use the following command to start the project 👇:\n\n")

	fmt.Println(
		color.WhiteString("$ cd %s", opts.Project.D.WorkDir),
		color.CyanString("# enter project directory"),
	)

	if opts.Project.Metadata.MakefileMode != known.MakefileModeNone {
		fmt.Println(
			color.WhiteString("$ make deps"),
			color.CyanString("# (Optional, executed when dependencies missing) Install tools required by project."),
		)

		switch n := len(opts.Project.WebServers); {
		case n == 1:
			fmt.Println(
				color.WhiteString("$ make protoc.%s", opts.Project.WebServers[0].Name),
				color.CyanString("# generate gRPC code"),
			)
		case n > 0:
			fmt.Println(
				color.WhiteString("$ make protoc.<component_name>"),
				color.CyanString("# generate gRPC code"),
			)
		}
	}

	fmt.Println(
		color.WhiteString("$ go get cloud.google.com/go/compute@latest"),
		color.CyanString("# to resolve `cloud.google.com/go/compute/metadata: ambiguous import`"),
	)
	fmt.Println(
		color.WhiteString("$ go get cloud.google.com/go/compute/metadata@latest"),
	)
	fmt.Println(
		color.WhiteString("$ go mod tidy"),
		color.CyanString("# tidy dependencies"),
	)
	fmt.Println(
		color.WhiteString("$ go generate ./..."),
		color.CyanString("# run all go:generate directives"),
	)

	if opts.Project.Metadata.MakefileMode == known.MakefileModeNone {
		PrintClosingTips(opts.Project.D.ProjectName)
		return
	}

	if len(opts.Project.WebServers) == 1 {
		ws := opts.Project.WebServers[0]
		if opts.Project.Metadata.MakefileMode != known.MakefileModeNone {
			fmt.Println(
				color.WhiteString("$ make build BINS=%s", ws.BinaryName),
				color.CyanString("# build %s", ws.BinaryName),
			)
		}
		if ws.StorageType == known.StorageTypeMemory {
			fmt.Println(
				color.WhiteString("$ _output/platforms/%s/%s/%s", runtime.GOOS, runtime.GOARCH, ws.BinaryName),
				color.CyanString("# run the compiled server"),
			)
			if ws.WebFramework == known.WebFrameworkGin && ws.WithHealthz {
				fmt.Println(
					color.WhiteString("$ curl http://127.0.0.1:5555/healthz"),
					color.CyanString("# test with the health endpoint"),
				)
			}
			if ws.WebFramework == known.WebFrameworkGRPC {
				if ws.WithUser {
					fmt.Println(
						color.WhiteString("$ go run examples/client/user/main.go"),
						color.CyanString("# run user client to test the API"),
					)
				} else if ws.WithHealthz {
					fmt.Println(
						color.WhiteString("$ go run examples/client/health/main.go"),
						color.CyanString("# run health client to test the API"),
					)
				}
			}
		}
	} else if len(opts.Project.WebServers) > 0 {
		fmt.Println(
			color.WhiteString("$ make build BINS=<binary-name>"),
			color.CyanString("# build the binary"),
		)
	}

	PrintClosingTips(opts.Project.D.ProjectName)
}

// Generate creates project-level files and per-component code based on templates.
func (opts *ProjectOptions) Generate(f cmdutil.Factory, fm *file.FileManager) error {
	// Persist project config into the root directory
	if err := opts.Project.Save(opts.Project.Join(known.ProjectFileName)); err != nil {
		return err
	}

	funcs := template.FuncMap{
		"kind":        helper.Kind(),
		"kinds":       helper.Kinds(),
		"capitalize":  helper.Capitalize(),
		"lowerkind":   helper.SingularLower(),
		"lowerkinds":  helper.SingularLowers(),
		"currentYear": helper.CurrentYear(),
	}

	// Project-level templates
	projectFiles := map[string]string{
		filepath.Join("go.mod"):                                     "/project/go.mod",
		filepath.Join("README.md"):                                  "/project/README.md",
		filepath.Join("scripts/boilerplate.txt"):                    "/project/scripts/boilerplate.txt",
		filepath.Join(".gitignore"):                                 "/project/gitignore.tpl",
		filepath.Join(".golangci.yaml"):                             "/project/golangci.yaml",
		filepath.Join(".protolint.yaml"):                            "/project/protolint.yaml",
		filepath.Join("docs/images/.keep"):                          "/keep.tpl",
		filepath.Join("docs/devel/en-US/.keep"):                     "/keep.tpl",
		filepath.Join("docs/devel/zh-CN/.keep"):                     "/keep.tpl",
		filepath.Join("docs/guide/en-US/.keep"):                     "/keep.tpl",
		filepath.Join("docs/guide/zh-CN/README.md"):                 "/project/docs/guide/zh-CN/README.md",
		filepath.Join("docs/guide/zh-CN/announcements.md"):          "/project/docs/guide/zh-CN/announcements.md",
		filepath.Join("docs/guide/zh-CN/introduction/README.md"):    "/project/docs/guide/zh-CN/introduction/README.md",
		filepath.Join("docs/guide/zh-CN/quickstart/README.md"):      "/project/docs/guide/zh-CN/quickstart/README.md",
		filepath.Join("docs/guide/zh-CN/installation/README.md"):    "/project/docs/guide/zh-CN/installation/README.md",
		filepath.Join("docs/guide/zh-CN/operation-guide/README.md"): "/project/docs/guide/zh-CN/operation-guide/README.md",
		filepath.Join("docs/guide/zh-CN/best-practice/README.md"):   "/project/docs/guide/zh-CN/best-practice/README.md",
		filepath.Join("docs/guide/zh-CN/faq/README.md"):             "/project/docs/guide/zh-CN/faq/README.md",
		// Scripts
		filepath.Join("scripts/coverage.awk"): "/project/scripts/coverage.awk",
	}

	// Deployment-specific files
	switch opts.Project.Metadata.DeploymentMethod {
	case known.DeploymentModeSystemd:
		projectFiles[filepath.Join("init", "README.md")] = "/project/init/README.md"
	}

	// Makefile
	switch opts.Project.Metadata.MakefileMode {
	case known.MakefileModeUnstructured:
		projectFiles["Makefile"] = "/project/Makefile.unstructed"
	case known.MakefileModeStructured:
		projectFiles["Makefile"] = "/project/Makefile.structed"
		projectFiles["scripts/make-rules/all.mk"] = "/project/scripts/make-rules/all.mk"
		projectFiles["scripts/make-rules/common.mk"] = "/project/scripts/make-rules/common.mk"
		projectFiles["scripts/make-rules/generate.mk"] = "/project/scripts/make-rules/generate.mk"
		projectFiles["scripts/make-rules/golang.mk"] = "/project/scripts/make-rules/golang.mk"
		projectFiles["scripts/make-rules/tools.mk"] = "/project/scripts/make-rules/tools.mk"
		projectFiles["scripts/make-rules/image.mk"] = "/project/scripts/make-rules/image.mk"
	default:
	}

	// Generate project-level files
	if err := Generate(fm, projectFiles, funcs, &types.TemplateData{Project: opts.Project}); err != nil {
		return err
	}

	// Generate per-webserver files
	for _, ws := range opts.Project.WebServers {
		completedWebServer := ws.Complete(opts.Project)
		componentFiles := completedWebServer.Pairs()
		data := types.TemplateData{Project: opts.Project, Web: completedWebServer}
		if err := Generate(fm, componentFiles, funcs, &data); err != nil {
			return err
		}
	}

	// TODO: Add jobs/CLI apps generation when templates are ready.
	return nil
}

// correctProjectConfig fixes inconsistent project configuration.
func correctProjectConfig(proj *types.Project) *types.Project {
	if proj.Metadata.DeploymentMethod == "" {
		proj.Metadata.DeploymentMethod = known.DeploymentModeDocker
	}
	if proj.Metadata.MakefileMode == "" {
		proj.Metadata.MakefileMode = known.MakefileModeUnstructured
	}
	if proj.Metadata.ShortDescription == "" {
		proj.Metadata.ShortDescription = "TODO: Update the short description of the binary file."
	}
	if proj.Metadata.LongMessage == "" {
		proj.Metadata.LongMessage = "TODO: Update the detailed description of the binary file."
	}
	if proj.Metadata.Image.RegistryPrefix == "" {
		proj.Metadata.Image.RegistryPrefix = "docker.io/_undefined"
	}
	if proj.Metadata.Image.DockerfileMode == "" {
		proj.Metadata.Image.DockerfileMode = known.DockerfileModeCombined
	}

	for _, ws := range proj.WebServers {
		if ws.WebFramework == "" {
			ws.WebFramework = known.WebFrameworkGin
		}
		if ws.StorageType == "" {
			ws.StorageType = known.StorageTypeMemory
		}
		if ws.WebFramework != known.WebFrameworkGRPC && ws.WebFramework != known.WebFrameworkGRPCGateway {
			ws.GRPCServiceName = ""
		}

		if ws.ServiceRegistry == "" {
			ws.ServiceRegistry = known.ServiceRegistryNone
		}
	}

	return proj
}

// validateMetadata validate metadata.
func validateMetadata(md *types.Metadata) error {
	// Validate deployment mode (project-level)
	dep := strings.TrimSpace(md.DeploymentMethod)
	if !known.AvailableDeploymentModes.Has(dep) {
		return fmt.Errorf(
			"unsupported metadata.deploymentMode %q; supported: %s",
			dep, strings.Join(known.AvailableDeploymentModes.UnsortedList(), ", "),
		)
	}

	// If use cloud-native deploy method, need to generate Dockerfile.
	if stringsutil.StringIn(md.DeploymentMethod, []string{known.DeploymentModeDocker, known.DeploymentModeKubernetes}) {
		if err := validateImageConfig(md.Image); err != nil {
			return err
		}
	}

	// Validate makefile mode (project-level)
	if !known.AvailableMakefileModes.Has(md.MakefileMode) {
		return fmt.Errorf(
			"unsupported metadata.makefileMode %q; supported: %s",
			md.MakefileMode,
			strings.Join(known.AvailableMakefileModes.UnsortedList(), ", "),
		)
	}

	if md.MakefileMode == known.MakefileModeNone {
		fmt.Println(color.YellowString("Warning! If `makefileMode` is set to none, you must manually execute commands to build the source code."))
	}

	return nil
}

func validateImageConfig(image types.ImageConfig) error {
	if image.RegistryPrefix == "" {
		return fmt.Errorf("metadata.image.registryPrefix cannot be empty")
	}

	// Validate dockerfile mode (project-level)
	dockerfileMode := strings.TrimSpace(image.DockerfileMode)
	if !known.AvailableDockerfileModes.Has(dockerfileMode) {
		return fmt.Errorf(
			"unsupported metadata.image.dockerfileMode %q; supported: %s",
			dockerfileMode, strings.Join(known.AvailableDockerfileModes.UnsortedList(), ", "),
		)
	}

	return nil
}
