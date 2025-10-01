package main

import (
	"os"

	"{{.D.ModuleName}}/cmd/{{.Web.BinaryName}}/app"
)

// The default entry point of a Go program. Serves as the starting point
// for reading the project code.
func main() {
	command := app.NewWebServerCommand()

	// Execute the command and handle errors.
	if err := command.Execute(); err != nil {
		// Exit the program if an error occurs.
		// Return an exit code so that other programs (e.g., bash scripts)
		// can determine the service status based on the exit code.
		os.Exit(1)
	}
}
