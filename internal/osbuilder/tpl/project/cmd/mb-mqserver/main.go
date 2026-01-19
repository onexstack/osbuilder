package main

import (
	"os"

	"{{.D.ModuleName}}/cmd/{{.Web.BinaryName}}/app"
)

// The default entry point of a Go program. Serves as the starting point
// for reading the project code.
func main() {
	// Initialize the main command for the apiserver application.
	command := app.NewAPIServerCommand()

	// Execute the command. If an error occurs, the program exits.
	// The exit code provides an indication of failure for external systems.
	if err := command.Execute(); err != nil {
		os.Exit(1)
	}
}
