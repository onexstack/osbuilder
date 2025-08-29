package create

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/fatih/color"
	"github.com/gobuffalo/flect"
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
	opts := NewAPIOptions(ioStreams)

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
			cmdutil.CheckErr(opts.Complete(factory, cmd, args))
			cmdutil.CheckErr(opts.Validate(cmd, args))
			cmdutil.CheckErr(opts.Run(args))
		},
	}

	// Flags
	cmd.Flags().StringSliceVarP(&opts.Kinds, "kinds", "", opts.Kinds, "Resource kinds to generate in snake_case (e.g., cron_job).")
	cmd.Flags().StringVarP(&opts.BinaryName, "binary-name", "b", opts.BinaryName, "Target binary/web server name (e.g., apiserver).")
	cmd.Flags().BoolVarP(&opts.Force, "force", "f", opts.Force, "Force overwriting of existing files.")
	cmd.Flags().BoolVar(&opts.ShowTips, "show-tips", opts.ShowTips, "Print post-run tips.")
	_ = cmd.Flags().MarkHidden("show-tips")

	return cmd
}

// Complete resolves working directory and loads project metadata.
func (opts *APIOptions) Complete(factory cmdutil.Factory, cmd *cobra.Command, args []string) error {
	wd, err := os.Getwd()
	if err != nil {
		return err
	}
	opts.RootDir = wd

	proj, err := LoadProjectFromFile(filepath.Join(opts.RootDir, known.ProjectFileName))
	if err != nil {
		return err
	}

	// Fill generated data
	proj.D = (&types.GeneratedData{
		WorkDir:    opts.RootDir,
		APIVersion: "v1",
		APIAlias:   "v1",
		ModuleName: MustModelName(opts.RootDir),
	}).Complete()
	proj.D.ProjectName = filepath.Base(opts.RootDir)
	proj.D.RegistryPrefix = filepath.Join(proj.Metadata.Registry, proj.D.ProjectName)

	// If a single web server exists and BinaryName not set, default to it.
	if opts.BinaryName == "" && len(proj.WebServers) == 1 {
		opts.BinaryName = proj.WebServers[0].Name
	}

	opts.Project = proj
	return nil
}

// Validate checks required inputs and project state.
func (opts *APIOptions) Validate(cmd *cobra.Command, args []string) error {
	if opts.Project == nil {
		return fmt.Errorf("project not loaded")
	}
	if len(opts.Kinds) == 0 {
		return fmt.Errorf("at least one kind must be provided via --kinds")
	}
	ws := opts.Project.FindWebServer(opts.BinaryName)
	if ws == nil {
		return fmt.Errorf("web server/binary %q not found in project; use --binary-name", opts.BinaryName)
	}
	if opts.APIVersion == "" {
		return fmt.Errorf("api version must not be empty")
	}
	return nil
}

// Run generates files for each kind and updates related components.
func (opts *APIOptions) Run(args []string) error {
	fm := file.NewFileManager(opts.RootDir, opts.Force)

	web := opts.Project.FindWebServer(opts.BinaryName).Complete(opts.Project)
	for _, kind := range opts.Kinds {
		// Build REST spec and attach to the selected web server
		web.SetREST(opts.BuildREST(kind))

		// Generate files (proto, handlers, validation, store, biz, model)
		if err := opts.GenerateFiles(fm, web); err != nil {
			return err
		}

		if web.WebFramework == known.WebFrameworkGRPC {
			// Update proto: append new gRPC service/methods and import
			importPath := filepath.Join(web.Name, opts.Project.D.APIVersion, web.R.SingularLower+".proto")
			protoFilePath := opts.Project.Join(web.API(), web.Name+".proto")
			if err := fm.AddNewGRPCMethod(protoFilePath, web.R.SingularName, web.GRPCServiceName, importPath); err != nil {
				return err
			}
		}

		// Update store.go
		internalDir := filepath.Join(opts.Project.D.WorkDir, fmt.Sprintf("internal/%s", web.Name))
		if err := fm.AddNewMethod("store", filepath.Join(internalDir, "store", "store.go"), web.R.SingularName, opts.Project.D.APIVersion, ""); err != nil {
			return err
		}

		// Update biz.go
		if err := fm.AddNewMethod(
			"biz",
			filepath.Join(internalDir, "biz", "biz.go"),
			web.R.SingularName,
			opts.Project.D.APIVersion,
			fmt.Sprintf("%s/internal/%s", opts.Project.D.ModuleName, web.Name),
		); err != nil {
			return err
		}
	}

	if opts.ShowTips {
		opts.PrintGettingStarted(web)
	}
	return nil
}

// BuildREST constructs REST metadata for a given kind.
func (opts *APIOptions) BuildREST(kind string) *types.REST {
	upperVer := strings.ToUpper(opts.Project.D.APIVersion)

	r := types.REST{
		SingularName:       helper.ToUpperCamelCase(kind),
		PluralName:         flect.Pluralize(helper.ToUpperCamelCase(kind)),
		SingularLowerFirst: helper.ToLowerCamelCase(kind),
		PluralLowerFirst:   flect.Pluralize(helper.ToLowerCamelCase(kind)),
	}

	r.SingularLower = strings.ToLower(r.SingularName)
	r.PluralLower = strings.ToLower(r.PluralName)
	r.GORMModel = r.SingularName + "M"
	r.MapModelToAPIFunc = fmt.Sprintf("%sMTo%s%s", r.SingularName, r.SingularName, upperVer)
	r.MapAPIToModelFunc = fmt.Sprintf("%s%sTo%sM", r.SingularName, upperVer, r.SingularName)
	r.BusinessFactoryName = fmt.Sprintf("%s%s", r.SingularName, upperVer)
	r.FileName = r.SingularLower + ".go"

	return &r
}

// GenerateFiles materializes files for the selected web server and kind.
func (opts *APIOptions) GenerateFiles(fm *file.FileManager, web *types.WebServer) error {
	pairs := map[string]string{
		filepath.Join(web.API(), web.R.SingularLower+".proto"):                                   "/project/pkg/api/apiserver/v1/post.proto",
		filepath.Join(web.Pkg(), "validation", web.R.FileName):                                   "/project/internal/apiserver/pkg/validation/post.go",
		filepath.Join(web.Store(), web.R.FileName):                                               "/project/internal/apiserver/store/post.go",
		filepath.Join(web.Biz(), opts.Project.D.APIVersion, web.R.SingularLower, web.R.FileName): "/project/internal/apiserver/biz/v1/post/post.go",
		filepath.Join(web.Model(), web.R.FileName):                                               "/project/internal/apiserver/model/post.gen.go",
		filepath.Join(web.Model(), "hook_"+web.R.FileName):                                       "/project/internal/apiserver/model/hook_post.go",
	}

	switch web.WebFramework {
	case known.WebFrameworkGin:
		pairs[filepath.Join(web.Handler(), web.R.FileName)] = "/project/internal/apiserver/handler/gin/post.go"
	case known.WebFrameworkGRPC:
		pairs[filepath.Join("examples/client", web.R.SingularLower, "main.go")] = "/project/examples/client/post/main.go"
		pairs[filepath.Join(web.Handler(), web.R.FileName)] = "/project/internal/apiserver/handler/grpc/post.go"
	}

	pairs[filepath.Join(web.Pkg(), "conversion", web.R.FileName)] = "/project/internal/apiserver/pkg/conversion/post.go"

	funcs := template.FuncMap{
		"kind":        helper.Kind(),
		"kinds":       helper.Kinds(),
		"capitalize":  helper.Capitalize(),
		"lowerkind":   helper.SingularLower(),
		"lowerkinds":  helper.SingularLowers(),
		"currentYear": helper.CurrentYear(),
	}

	// Generate templated files using the provided template engine
	if err := Generate(fm, pairs, funcs, &types.TemplateData{Project: opts.Project, Web: web}); err != nil {
		return err
	}
	return nil
}

// PrintGettingStarted prints follow-up commands to rebuild and generate gRPC assets.
func (opts *APIOptions) PrintGettingStarted(web *types.WebServer) {
	fmt.Printf("\n🍺 REST resources creation succeeded %s\n", color.GreenString("%s", strings.Join(opts.Kinds, ",")))
	fmt.Print("💻 Use the following command to re-compile the project 👇:\n\n")

	fmt.Println(
		color.WhiteString("$ cd %s", opts.RootDir),
		color.CyanString("# enter project directory"),
	)
	fmt.Println(
		color.WhiteString("$ make protoc.%s", web.Name),
		color.CyanString("# generate gRPC code"),
	)
	fmt.Println(
		color.WhiteString("$ make build BINS=%s", web.BinaryName),
		color.CyanString("# build %s", web.BinaryName),
	)
	fmt.Println(color.WhiteString("After restarting, you can run `go run examples/client/<kind>/main.go` to test the new resource."))

	PrintClosingTips(opts.Project.D.ProjectName)
}
