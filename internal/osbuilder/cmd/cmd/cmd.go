package cmd

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
)

// CmdOptions defines configuration for the "create cmd" generator.
type CmdOptions struct {
	RootDir string // working directory
	Name    string // name of the new subcommand
	Output  string // output directory
	Force   bool   // overwrite if exists

	genericiooptions.IOStreams
}

type TemplateData struct {
	Output          string
	StructName      string
	PackageName     string
	CommandName     string
	ConstructorName string
}

var (
	cmdLongDesc = templates.LongDesc(`
        Generate a new cobra sub-command under the specified parent command.

        It scaffolds a new Go source file including command declaration, example flags, and help text.
    `)

	cmdExample = templates.Examples(`
        # Create a new 'hello' subcommand under the root command
        osbuilder create cmd --name hello

        # Create a subcommand under an existing parent command 'create'
        osbuilder create cmd --name plugin --parent create

        # Specify output directory
        osbuilder create cmd --name foo --output internal/osbuilder/cmd/util
    `)
)

// NewCmdCmd defines the "create cmd" command.
func NewCmdCmd(factory cmdutil.Factory, ioStreams genericiooptions.IOStreams) *cobra.Command {
	opts := &CmdOptions{
		IOStreams: ioStreams,
		Output:    "./", // default output directory
	}

	cmd := &cobra.Command{
		Use:                   "cmd",
		Short:                 "Create a new cobra sub-command",
		Long:                  cmdLongDesc,
		Example:               cmdExample,
		SilenceUsage:          true,
		SilenceErrors:         true,
		DisableFlagsInUseLine: true,
		Run: func(cmd *cobra.Command, args []string) {
			cmdutil.CheckErr(opts.Complete())
			cmdutil.CheckErr(opts.Validate())
			cmdutil.CheckErr(opts.Run())
		},
	}

	cmd.Flags().StringVarP(&opts.Name, "name", "n", "", "Name of the new cobra subcommand (required).")
	cmd.Flags().StringVarP(&opts.Output, "output", "o", opts.Output, "Output directory for generated command file.")
	cmd.Flags().BoolVarP(&opts.Force, "force", "f", false, "Overwrite existing file if it exists.")

	return cmd
}

// Complete sets default paths and resolves working directory.
func (o *CmdOptions) Complete() error {
	o.RootDir, _ = os.Getwd()
	return nil
}

// Validate ensures provided inputs are valid.
func (o *CmdOptions) Validate() error {
	if o.Name == "" {
		return fmt.Errorf("--name is required")
	}
	if strings.Contains(o.Name, " ") {
		return fmt.Errorf("--name cannot contain spaces")
	}
	if strings.Contains(o.Name, "/") || strings.Contains(o.Name, "\\") {
		return fmt.Errorf("--name cannot contain path separators")
	}
	return nil
}

// Run performs the subcommand file generation.
func (o *CmdOptions) Run() error {
	fm := file.NewFileManager(o.RootDir, o.Force)

	pairs := map[string]string{o.Name + ".go": "/cmd/cmd.go"}

	td := &TemplateData{
		Output:          o.Output,
		StructName:      helper.ToUpperCamelCase(strings.ToLower(o.Name)),
		PackageName:     strings.ToLower(helper.ToUpperCamelCase(o.Name)),
		CommandName:     strings.ToLower(o.Name),
		ConstructorName: fmt.Sprintf("New%sCmd", helper.ToUpperCamelCase(strings.ToLower(o.Name))),
	}

	// Generate the command file using template
	if err := helper.RenderTemplate(fm, pairs, helper.GetTemplateFuncMap(), td); err != nil {
		return fmt.Errorf("failed to render template: %w", err)
	}

	// Print success message
	o.PrintGettingStarted()

	return nil
}

// AbsPath  the absolute path of the generated command file.
func (o *CmdOptions) AbsPath() string {
	return filepath.Join(o.Output, o.Name+".go")
}

// AbsPath implements TemplateDataProvider interface.
func (td *TemplateData) AbsPath(relPath string) string {
	return filepath.Join(td.Output, relPath)
}

// PrintGettingStarted() prints formatted success information.
func (o *CmdOptions) PrintGettingStarted() {
	fmt.Printf("%s Successfully created new cobra subcommand: %s\n", emoji.CheckMarkButton, color.GreenString(o.Name))
	fmt.Printf("%s Location: %s\n", emoji.PageFacingUp, color.CyanString(o.AbsPath()))
	fmt.Printf("%s Tip: Add it to your parent command using:\n", emoji.LightBulb)
	fmt.Printf("     parentCmd.AddCommand(New%sCmd())\n\n", helper.ToUpperCamelCase(o.Name))
}
