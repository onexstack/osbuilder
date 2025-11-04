package beforebump

import (
	"github.com/onexstack/osbuilder/internal/osbuilder/semver/context"
	"github.com/onexstack/osbuilder/internal/osbuilder/semver/task/hook"
)

// Task for executing any custom shell commands or scripts
// before file bumping within the release workflow
type Task struct{}

// String generates a string representation of the task
func (t Task) String() string {
	return "before bumping files"
}

// Skip running the task
func (t Task) Skip(ctx *context.Context) bool {
	return ctx.Config.Hooks == nil ||
		len(ctx.Config.Hooks.BeforeBump) == 0 ||
		ctx.NoVersionChanged ||
		ctx.SkipBumps
}

// Run the task
func (t Task) Run(ctx *context.Context) error {
	return hook.Exec(ctx.Context, ctx.Config.Hooks.BeforeBump, hook.ExecOptions{
		DryRun: ctx.DryRun,
		Debug:  ctx.Debug,
		Env:    ctx.Config.Env,
	})
}
