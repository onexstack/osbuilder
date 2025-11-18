package create

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/enescakir/emoji"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"k8s.io/cli-runtime/pkg/genericiooptions"
	"k8s.io/kubectl/pkg/util/templates"

	cmdutil "github.com/onexstack/osbuilder/internal/osbuilder/cmd/util"
	"github.com/onexstack/osbuilder/internal/osbuilder/file"
	"github.com/onexstack/osbuilder/internal/osbuilder/helper"
	"github.com/onexstack/osbuilder/internal/osbuilder/known"
	"github.com/onexstack/osbuilder/internal/osbuilder/types"
)

// APIOptions holds flags and runtime context for the 'create api' command.
type APIOptions struct {
	RootDir string

	Kinds      []string // Resource kinds to generate (snake_case recommended)
	BinaryName string   // Target web server/binary name
	Force      bool     // Overwrite files if they exist

	APIVersion string // API version, e.g., "v1"
	ShowTips   bool   // Print getting-started hints

	Project *types.Project // Loaded project metadata

	genericiooptions.IOStreams
}

var (
	apiLongDesc = templates.LongDesc(`
		Create API resources for your project.

		This command scaffolds API artifacts (proto, handlers, validation, store, biz, model) for the given kinds.`)

	apiExamples = templates.Examples(`
		# Create API resources for a specific kind
		osbuilder create api --kinds post --binary-name mb-apiserver

		# Create multiple kinds
		osbuilder create api --kinds cron_job,job --binary-name mb-apiserver`)
)

// NewAPIOptions creates a default APIOptions.
func NewAPIOptions(io genericiooptions.IOStreams) *APIOptions {
	return &APIOptions{
		APIVersion: "v1",
		ShowTips:   true,
		IOStreams:  io,
	}
}

// NewCmdAPI builds the 'create api' cobra command.
func NewCmdAPI(factory cmdutil.Factory, ioStreams genericiooptions.IOStreams) *cobra.Command {
	o := NewAPIOptions(ioStreams)

	cmd := &cobra.Command{
		Use:                   "api",
		DisableFlagsInUseLine: true,
		Short:                 "Create API resources",
		Long:                  apiLongDesc,
		Example:               apiExamples,
		SilenceUsage:          true,
		SilenceErrors:         true,
		Args:                  cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			cmdutil.CheckErr(o.Complete(factory, cmd, args))
			cmdutil.CheckErr(o.Validate(cmd, args))
			cmdutil.CheckErr(o.Run(args))
		},
	}

	// Flags
	cmd.Flags().StringSliceVarP(&o.Kinds, "kinds", "", o.Kinds, "Resource kinds to generate in snake_case (e.g., cron_job).")
	cmd.Flags().StringVarP(&o.BinaryName, "binary-name", "b", o.BinaryName, "Target binary/web server name (e.g., mb-apiserver).")
	cmd.Flags().BoolVarP(&o.Force, "force", "f", o.Force, "Force overwriting of existing files.")
	// Add hidden flags
	cmd.Flags().StringVar(&o.RootDir, "root-dir", "", "Override root directory (hidden flag)")
	cmd.Flags().BoolVar(&o.ShowTips, "show-tips", o.ShowTips, "Print post-run tips.")
	_ = cmd.Flags().MarkHidden("root-dir")
	_ = cmd.Flags().MarkHidden("show-tips")

	return cmd
}

// Complete resolves working directory and loads project metadata.
func (o *APIOptions) Complete(factory cmdutil.Factory, cmd *cobra.Command, args []string) error {
	if o.RootDir == "" {
		wd, err := os.Getwd()
		if err != nil {
			return err
		}
		o.RootDir = wd
	}

	proj, err := LoadProjectFromFile(filepath.Join(o.RootDir, known.ProjectFileName))
	if err != nil {
		return err
	}

	// Fill generated data
	proj.D = (&types.GeneratedData{
		WorkDir:    o.RootDir,
		APIVersion: "v1",
		APIAlias:   "v1",
		ModuleName: MustModulePath(proj.Metadata.ModulePath, o.RootDir),
	}).Complete()
	proj.D.ProjectName = filepath.Base(o.RootDir)
	proj.D.RegistryPrefix = proj.Metadata.Image.RegistryPrefix

	// If a single web server exists and BinaryName not set, default to it.
	if o.BinaryName == "" && len(proj.WebServers) == 1 {
		o.BinaryName = proj.WebServers[0].Name
	}

	o.Project = proj
	return nil
}

// Validate checks required inputs and project state.
func (o *APIOptions) Validate(cmd *cobra.Command, args []string) error {
	if o.Project == nil {
		return fmt.Errorf("project not loaded")
	}
	if len(o.Kinds) == 0 {
		return fmt.Errorf("at least one kind must be provided via --kinds")
	}
	ws := o.Project.FindWebServer(o.BinaryName)
	if ws == nil {
		return fmt.Errorf("web server/binary %q not found in project; use --binary-name", o.BinaryName)
	}
	if o.APIVersion == "" {
		return fmt.Errorf("api version must not be empty")
	}
	return nil
}

// Run generates files for each kind and updates related components.
func (o *APIOptions) Run(args []string) (err error) {
	defer func() { helper.RecordOSBuilderUsage("api", err) }()

	fm := file.NewFileManager(o.RootDir, o.Force)

	ws := o.Project.FindWebServer(o.BinaryName).Complete(o.Project)
	for _, kind := range o.Kinds {
		// Build REST spec and attach to the selected web server
		ws.PrepareRESTMetadata(kind)

		// Generate files (proto, handlers, validation, store, biz, model)
		if err := o.GenerateFiles(fm, ws); err != nil {
			return err
		}

		if ws.WebFramework == known.WebFrameworkGRPC {
			// Update proto: append new gRPC service/methods and import
			importPath := filepath.Join(ws.Name, o.Project.D.APIVersion, ws.R.SingularLower+".proto")
			protoFilePath := o.Project.Join(ws.API(), ws.Name+".proto")
			if err := fm.AddNewGRPCMethod(protoFilePath, ws.R.SingularName, ws.GRPCServiceName, importPath); err != nil {
				return err
			}
		}

		// Update store.go
		internalDir := filepath.Join(o.Project.D.WorkDir, fmt.Sprintf("internal/%s", ws.Name))
		if err := fm.AddNewMethod("store", filepath.Join(internalDir, "store", "store.go"), ws.R.SingularName, o.Project.D.APIVersion, ""); err != nil {
			return err
		}

		// Update biz.go
		if err := fm.AddNewMethod(
			"biz",
			filepath.Join(internalDir, "biz", "biz.go"),
			ws.R.SingularName,
			o.Project.D.APIVersion,
			fmt.Sprintf("%s/internal/%s/biz/%s/%s/%s",
				o.Project.D.ModuleName,
				ws.Name,
				o.Project.D.APIVersion,
				ws.R.ResourceDirPrefix,
				ws.R.SingularLower,
			),
		); err != nil {
			return err
		}
	}

	if o.ShowTips {
		o.PrintGettingStarted(ws)
	}
	return nil
}

// GenerateFiles materializes files for the selected web server and kind.
func (o *APIOptions) GenerateFiles(fm *file.FileManager, ws *types.WebServer) error {
	pairs := map[string]string{
		filepath.Join(ws.API(), ws.R.SingularLower+".proto"):         "/project/pkg/api/apiserver/v1/post.proto",
		filepath.Join(ws.Pkg(), "validation", ws.R.FileName):         "/project/internal/apiserver/pkg/validation/post.go",
		filepath.Join(ws.Store(), ws.R.FileName):                     "/project/internal/apiserver/store/post.go",
		filepath.Join(ws.RESTBiz()):                                  "/project/internal/apiserver/biz/v1/post/post.go",
		filepath.Join(ws.Model(), ws.R.FileName):                     "/project/internal/apiserver/model/post.gen.go",
		filepath.Join(ws.Model(), "hook_"+ws.R.FileName):             "/project/internal/apiserver/model/hook_post.go",
		filepath.Join(ws.Proj.InternalPkg(), "errno", ws.R.FileName): "/project/internal/pkg/errno/post.go",
	}

	switch ws.WebFramework {
	case known.WebFrameworkGin:
		pairs[filepath.Join(ws.Handler(), ws.R.FileName)] = "/project/internal/apiserver/handler/gin/post.go"
	case known.WebFrameworkGRPC:
		pairs[filepath.Join("examples/client", ws.R.SingularLower, "main.go")] = "/project/examples/client/post/main.go"
		pairs[filepath.Join(ws.Handler(), ws.R.FileName)] = "/project/internal/apiserver/handler/grpc/post.go"
	}

	pairs[filepath.Join(ws.Pkg(), "conversion", ws.R.FileName)] = "/project/internal/apiserver/pkg/conversion/post.go"

	// Generate templated files using the provided template engine
	if err := helper.RenderTemplate(
		fm,
		pairs,
		helper.GetTemplateFuncMap(),
		&types.TemplateData{Project: o.Project, Web: ws},
	); err != nil {
		return err
	}
	return nil
}

// PrintGettingStarted prints follow-up commands to rebuild and generate gRPC assets.
func (o *APIOptions) PrintGettingStarted(ws *types.WebServer) {
	fmt.Printf("\n%s REST resource(s) creation succeeded %s\n", emoji.CheckMarkButton, color.GreenString("%s", strings.Join(o.Kinds, ",")))
	if o.Project.Metadata.MakefileMode == known.MakefileModeNone {
		PrintClosingTips(o.Project.D.ProjectName)
		return
	}

	fmt.Printf("%s Use the following command to re-compile the project %s:\n\n", emoji.Parse(":computer:"), emoji.Parse(":point_down:"))

	fmt.Println(
		color.WhiteString("$ cd %s", o.RootDir),
		color.CyanString("# enter project directory"),
	)
	fmt.Println(
		color.WhiteString("$ make protoc.%s", ws.Name),
		color.CyanString("# generate gRPC code"),
	)
	fmt.Println(
		color.WhiteString("$ make build BINS=%s", ws.BinaryName),
		color.CyanString("# build %s", ws.BinaryName),
	)
	fmt.Println(color.WhiteString("After restarting, you can run `go run examples/client/<kind>/main.go` to test the new resource."))

	PrintClosingTips(o.Project.D.ProjectName)
}
