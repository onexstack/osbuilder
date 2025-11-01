package create

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
	"k8s.io/cli-runtime/pkg/genericiooptions"
	"k8s.io/kubectl/pkg/util/templates"

	cmdutil "github.com/onexstack/osbuilder/internal/osbuilder/cmd/util"
	//"github.com/onexstack/osbuilder/internal/osbuilder/file"
	"github.com/onexstack/osbuilder/internal/osbuilder/helper"
	"github.com/onexstack/osbuilder/internal/osbuilder/types"
)

// QuickstartOptions holds flags and runtime context for the 'create quickstart' command.
type QuickstartOptions struct {
	ProjectRootDir string

	// Project configuration parameters
	ModuleName  string // Go module name
	ProjectName string // Project name

	Author         string // Project author
	Email          string // Author email
	MakefileMode   string // Makefile mode
	DeploymentMode string // Deployment mode
	RegistryPrefix string // Container registry prefix
	Distroless     bool   // Use distroless base image

	BinaryName      string   // Target web server/binary name
	Kinds           []string // Resource kinds to generate (snake_case recommended)
	WebFramework    string   // Web framework to use
	WithUser        bool     // Include user management logic
	WithOtel        bool     // Enable OpenTelemetry
	ServiceRegistry string   // Service registry type

	genericiooptions.IOStreams
}

var (
	quickstartLongDesc = templates.LongDesc(`
		Create a quickstart demo project with sensible defaults.

		This command scaffolds a complete project structure with configurable options
		for web framework, deployment mode, and other common settings. It's designed
		to help you get started quickly with a working project template.`)

	quickstartExamples = templates.Examples(`
		# Create a quickstart project with default settings
		osbuilder create quickstart ./my-demo

		# Create with custom module name and web framework
		osbuilder create quickstart ./my-project \
		  --module-name github.com/myorg/myproject \
		  --web-framework grpc

		# Create with full customization
		osbuilder create quickstart ./enterprise-api \
		  --module-name github.com/company/enterprise-api \
		  --project-name enterprise-api \
		  --author "孔令飞" \
		  --email colin404@foxmail.com \
		  --makefile-mode structured \
		  --deployment-mode kubernetes \
		  --registry-prefix docker.io/company \
		  --distroless \
		  --binary-name os-apiserver \
		  --kinds cron_job,job \
		  --web-framework gin \
		  --with-user \
		  --with-otel \
		  --service-registry polaris`)
)

// NewQuickstartOptions creates a default QuickstartOptions.
func NewQuickstartOptions(io genericiooptions.IOStreams) *QuickstartOptions {
	return &QuickstartOptions{
		ModuleName:      "",
		ProjectName:     "osdemo",
		Author:          "孔令飞",
		Email:           "colin404@foxmail.com",
		MakefileMode:    "unstructured",
		DeploymentMode:  "docker",
		RegistryPrefix:  "", // Will be set to docker.io/<project-name> if empty
		Distroless:      false,
		BinaryName:      "os-apiserver",
		Kinds:           []string{"cron_job", "job"},
		WebFramework:    "gin",
		WithUser:        false,
		WithOtel:        true,
		ServiceRegistry: "none",
		IOStreams:       io,
	}
}

// NewCmdQuickstart builds the 'create quickstart' cobra command.
func NewCmdQuickstart(factory cmdutil.Factory, ioStreams genericiooptions.IOStreams) *cobra.Command {
	opts := NewQuickstartOptions(ioStreams)

	cmd := &cobra.Command{
		Use:                   "quickstart [PROJECT_NAME]",
		DisableFlagsInUseLine: true,
		Short:                 "Create a quickstart demo project",
		Long:                  quickstartLongDesc,
		Example:               quickstartExamples,
		SilenceUsage:          true,
		SilenceErrors:         true,
		Args:                  cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			cmdutil.CheckErr(opts.Complete(factory, cmd, args))
			cmdutil.CheckErr(opts.Validate(cmd, args))
			cmdutil.CheckErr(opts.Run(factory, ioStreams, args))
		},
	}

	// Project configuration flags
	cmd.Flags().StringVar(&opts.ModuleName, "module-name", opts.ModuleName, "Go module name for the project. Default: onexstack/onexstack/<PROJECT_NAME>")
	cmd.Flags().StringVar(&opts.ProjectName, "project-name", opts.ProjectName, "Project name")
	cmd.Flags().StringVar(&opts.Author, "author", opts.Author, "Project author name")
	cmd.Flags().StringVar(&opts.Email, "email", opts.Email, "Author email address")
	cmd.Flags().StringVar(&opts.MakefileMode, "makefile-mode", opts.MakefileMode, "Makefile mode (none, unstructured, structured)")
	cmd.Flags().StringVar(&opts.DeploymentMode, "deployment-mode", opts.DeploymentMode, "Deployment mode (docker, kubernetes, systemd)")
	cmd.Flags().StringVar(&opts.RegistryPrefix, "registry-prefix", opts.RegistryPrefix, "Container registry prefix (default: docker.io/<project-name>)")
	cmd.Flags().BoolVar(&opts.Distroless, "distroless", opts.Distroless, "Use distroless base image for containers")
	cmd.Flags().StringVar(&opts.BinaryName, "binary-name", opts.BinaryName, "Target binary/web server name (e.g., os-apiserver).")
	cmd.Flags().StringSliceVarP(&opts.Kinds, "kinds", "", opts.Kinds, "Resource kinds to generate in snake_case (e.g., cron_job).")
	cmd.Flags().StringVar(&opts.WebFramework, "web-framework", opts.WebFramework, "Web framework to use (gin, grpc, grpc-gateway)")
	cmd.Flags().BoolVar(&opts.WithUser, "with-user", opts.WithUser, "Include user management, authentication and authorization logic")
	cmd.Flags().BoolVar(&opts.WithOtel, "with-otel", opts.WithOtel, "Enable OpenTelemetry support")
	cmd.Flags().StringVar(&opts.ServiceRegistry, "service-registry", opts.ServiceRegistry, "Service registry type (none, etcd, consul)")

	return cmd
}

// Complete resolves working directory and builds project configuration.
func (opts *QuickstartOptions) Complete(factory cmdutil.Factory, cmd *cobra.Command, args []string) error {
	if opts.ModuleName == "" {
		opts.ModuleName = "github.com/onexstack/" + opts.ProjectName
	}

	// Set default registry prefix if not provided
	if opts.RegistryPrefix == "" {
		opts.RegistryPrefix = fmt.Sprintf("docker.io/%s", opts.ProjectName)
	}

	wd, err := os.Getwd()
	if err != nil {
		return err
	}
	opts.ProjectRootDir = filepath.Join(wd, opts.ProjectName)
	return nil
}

// Validate checks required inputs and validates configuration options.
func (opts *QuickstartOptions) Validate(cmd *cobra.Command, args []string) error {
	return nil
}

// Run generates the quickstart project files and prints next steps.
func (opts *QuickstartOptions) Run(f cmdutil.Factory, ioStreams genericiooptions.IOStreams, args []string) (err error) {
	defer func() { helper.RecordOSBuilderUsage("project", err) }()

	fmt.Printf("\n🍺 Creating quickstart project %s...\n", color.GreenString(opts.ProjectName))
	projectString := helper.NewFileSystem("/").Content("/project.yaml")
	if projectString == "" {
		return fmt.Errorf("project template not found")
	}

	// 创建 reader 并解析 YAML
	reader := strings.NewReader(projectString)
	proj, err := DecodeProjectYAML(reader, true)
	if err != nil {
		return fmt.Errorf("decode project template: %w", err)
	}

	modifiedProject := opts.applyQuickstartOptions(proj)

	yamlString, err := projectToYAMLString(modifiedProject)
	if err != nil {
		return fmt.Errorf("convert project to yaml: %w", err)
	}

	encodedString := base64.StdEncoding.EncodeToString([]byte(yamlString))

	projectCmd := NewCmdProject(f, ioStreams)
	projectCmd.SetArgs([]string{opts.ProjectName, "--config-base64", encodedString, "--show-tips=false"})

	// 执行命令
	if err := projectCmd.Execute(); err != nil {
		fmt.Printf("Error executing project command: %v\n", err)
	}

	apiCmd := NewCmdAPI(f, ioStreams)
	apiCmd.SetArgs([]string{
		"--root-dir", opts.ProjectRootDir,
		"--binary-name", opts.BinaryName,
		"--kinds", strings.Join(opts.Kinds, ","),
		"--show-tips=false",
	})

	// 执行命令
	if err := apiCmd.Execute(); err != nil {
		fmt.Printf("Error executing api command: %v\n", err)
	}

	projectOptions := NewProjectOptions(ioStreams)
	projectOptions.ConfigBase64 = encodedString
	_ = projectOptions.Complete(f, nil, []string{opts.ProjectName})
	projectOptions.PrintGettingStarted()
	return nil
}

// BuildProject constructs a project configuration from command-line options.
func (opts *QuickstartOptions) BuildProject() *types.Project {
	return nil
}

// applyQuickstartOptions 根据 quickstart 命令行参数修改项目配置
func (opts *QuickstartOptions) applyQuickstartOptions(project *types.Project) *types.Project {
	// 修改 metadata 配置
	project.Metadata.ModulePath = opts.ModuleName

	if opts.Author != "" {
		project.Metadata.Author = opts.Author
	}
	if opts.Email != "" {
		project.Metadata.Email = opts.Email
	}
	if opts.MakefileMode != "" {
		project.Metadata.MakefileMode = opts.MakefileMode
	}
	if opts.DeploymentMode != "" {
		project.Metadata.DeploymentMethod = opts.DeploymentMode
	}
	if opts.RegistryPrefix != "" {
		project.Metadata.Image.RegistryPrefix = opts.RegistryPrefix
	}
	project.Metadata.Image.Distroless = opts.Distroless

	// 修改 WebServers 配置
	for i := range project.WebServers {
		if opts.BinaryName != "" {
			project.WebServers[i].BinaryName = opts.BinaryName
		}
		if opts.WebFramework != "" {
			project.WebServers[i].WebFramework = opts.WebFramework
		}
		project.WebServers[i].WithUser = opts.WithUser
		project.WebServers[i].WithOTel = opts.WithOtel
		if opts.ServiceRegistry != "" {
			project.WebServers[i].ServiceRegistry = opts.ServiceRegistry
		}
	}

	return project
}

// projectToYAMLString 将项目配置转换为 YAML 格式字符串
func projectToYAMLString(project *types.Project) (string, error) {
	if project == nil {
		return "", fmt.Errorf("project is nil")
	}

	var buf bytes.Buffer
	encoder := yaml.NewEncoder(&buf)
	encoder.SetIndent(2) // 设置缩进为 2 个空格

	if err := encoder.Encode(project); err != nil {
		return "", fmt.Errorf("encode project to yaml: %w", err)
	}

	if err := encoder.Close(); err != nil {
		return "", fmt.Errorf("close yaml encoder: %w", err)
	}

	return buf.String(), nil
}
