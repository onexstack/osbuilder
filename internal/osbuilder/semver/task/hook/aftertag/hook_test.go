package aftertag

import (
	"testing"

	"github.com/onexstack/osbuilder/internal/osbuilder/semver/config"
	"github.com/onexstack/osbuilder/internal/osbuilder/semver/context"
	"github.com/purpleclay/gitz/gittest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestString(t *testing.T) {
	assert.Equal(t, "after tagging repository", Task{}.String())
}

func TestSkip(t *testing.T) {
	assert.True(t, Task{}.Skip(&context.Context{
		Config: config.Uplift{
			Hooks: &config.Hooks{
				AfterTag: []string{},
			},
		},
	}))
}

func TestSkip_NoVersionChanged(t *testing.T) {
	assert.True(t, Task{}.Skip(&context.Context{
		NoVersionChanged: true,
	}))
}

func TestRun(t *testing.T) {
	// git.MkTmpDir(t)
	gittest.InitRepository(t)

	err := Task{}.Run(&context.Context{
		Config: config.Uplift{
			Hooks: &config.Hooks{
				AfterTag: []string{"touch a.out"},
			},
		},
	})
	require.NoError(t, err)
	assert.FileExists(t, "a.out")
}
