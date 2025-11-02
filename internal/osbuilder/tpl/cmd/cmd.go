package {{.PackageName}}

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/enescakir/emoji"
	"github.com/spf13/cobra"
	"k8s.io/cli-runtime/pkg/genericiooptions"
	"k8s.io/kubectl/pkg/util/templates"

	cmdutil "github.com/onexstack/osbuilder/internal/osbuilder/cmd/util"
)

// {{.StructName}}Options defines configuration for the "{{.CommandName}}" command.
type {{.StructName}}Options struct {
	RootDir string // working directory
	Output  string // output directory
	Force   bool   // overwrite if exists

	genericiooptions.IOStreams
}

var (
	{{.CommandName}}LongDesc = templates.LongDesc(`{{.CommandName}} command provides functionality for...

		Add your detailed description here.`)

	{{.CommandName}}Example = templates.Examples(`# Basic usage
		osbuilder {{.CommandName}} [options]

		# With custom options
		osbuilder {{.CommandName}} --output ./custom/path`)
)

// {{.ConstructorName}} creates the "{{.CommandName}}" command.
func {{.ConstructorName}}(factory cmdutil.Factory, ioStreams genericiooptions.IOStreams) *cobra.Command {
	opts := &{{.StructName}}Options{
		IOStreams: ioStreams,
		Output:    "./",
	}

	cmd := &cobra.Command{
		Use:                   "{{.CommandName}}",
		Short:                 "{{.CommandName}} command description",
		Long:                  {{.CommandName}}LongDesc,
		Example:               {{.CommandName}}Example,
		SilenceUsage:          true,
		SilenceErrors:         true,
		DisableFlagsInUseLine: true,
		Run: func(cmd *cobra.Command, args []string) {
			cmdutil.CheckErr(opts.Complete())
			cmdutil.CheckErr(opts.Validate())
			cmdutil.CheckErr(opts.Run())
		},
	}

	// Add flags
	cmd.Flags().StringVar(&opts.Output, "output", opts.Output, "Output directory for generated files.")
	cmd.Flags().BoolVarP(&opts.Force, "force", "f", false, "Overwrite existing files if they exist.")

	return cmd
}

// Complete sets default values and resolves working directory.
func (o *{{.StructName}}Options) Complete() error {
	o.RootDir, _ = os.Getwd()
	return nil
}

// Validate ensures provided inputs are valid.
func (o *{{.StructName}}Options) Validate() error {
	// Add your validation logic here
	return nil
}

// Run performs the {{.CommandName}} operation.
func (o *{{.StructName}}Options) Run() error {
	// TODO: Implement your command logic here
	o.PrintGettingStarted()
	return nil
}

// PrintGettingStarted prints formatted success information.
func (o *{{.StructName}}Options) PrintGettingStarted() {
	fmt.Printf("%s Successfully executed {{.CommandName}} command\n", emoji.CheckMarkButton)
}
