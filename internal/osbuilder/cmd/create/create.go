package create

import (
	"github.com/spf13/cobra"
	"k8s.io/cli-runtime/pkg/genericiooptions"
	"k8s.io/kubectl/pkg/util/templates"

	cmdutil "github.com/onexstack/osbuilder/internal/osbuilder/cmd/util"
)

// Long description for the 'create' command.
// Improves clarity, grammar, and intent.
var createLongDesc = templates.LongDesc(`
    Create a new project using the onexstack layout.

    This command serves as a root for project scaffolding and exposes subcommands 
    to create a full project or add an APIServer service to an existing directory.
`)

// NewCmdCreate returns the root 'create' command with its subcommands.
func NewCmdCreate(f cmdutil.Factory, ioStreams genericiooptions.IOStreams) *cobra.Command {
	cmd := &cobra.Command{
		Use:                   "create [command]",
		DisableFlagsInUseLine: true,
		Short:                 "Create a new project",
		Long:                  createLongDesc,
		SilenceUsage:          true, // Do not print usage on user errors
		// Show help when no subcommand is provided.
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	}

	// Allow visiting child flags without requiring explicit subcommand.
	cmd.TraverseChildren = true

	// Add subcommands.
	cmd.AddCommand(NewCmdProject(f, ioStreams))
	// Add an APIServer service within an existing directory.
	cmd.AddCommand(NewCmdAPI(f, ioStreams))

	return cmd
}
