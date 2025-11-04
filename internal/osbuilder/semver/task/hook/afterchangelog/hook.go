package afterchangelog

import (
	"github.com/onexstack/osbuilder/internal/osbuilder/semver/context"
	"github.com/onexstack/osbuilder/internal/osbuilder/semver/task/hook"
)

// Task for executing any custom shell commands or scripts
// after changelog generation within the release workflow
type Task struct{}

// String generates a string representation of the task
func (t Task) String() string {
	return "after generating changelog"
}

// Skip running the task
func (t Task) Skip(ctx *context.Context) bool {
	return ctx.Config.Hooks == nil ||
		len(ctx.Config.Hooks.AfterChangelog) == 0 ||
		ctx.SkipChangelog ||
		ctx.NoVersionChanged
}

// Run the task
func (t Task) Run(ctx *context.Context) error {
	return hook.Exec(ctx.Context, ctx.Config.Hooks.AfterChangelog, hook.ExecOptions{
		DryRun: ctx.DryRun,
		Debug:  ctx.Debug,
		Env:    ctx.Config.Env,
	})
}
