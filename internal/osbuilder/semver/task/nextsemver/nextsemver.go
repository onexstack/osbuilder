package nextsemver

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	semv "github.com/Masterminds/semver"
	"github.com/apex/log"
	git "github.com/purpleclay/gitz"

	"github.com/onexstack/osbuilder/internal/osbuilder/semver/context"
	"github.com/onexstack/osbuilder/internal/osbuilder/semver/semver"
)

// Task that determines the next semantic version of a repository
// based on the conventional commit of the last commit message
type Task struct{}

// String generates a string representation of the task
func (t Task) String() string {
	return "next semantic version"
}

// Skip is disabled for this task
func (t Task) Skip(_ *context.Context) bool {
	return false
}

// Run the task
func (t Task) Run(ctx *context.Context) error {
	var tagSuffix string
	if ctx.FilterOnPrerelease {
		tagSuffix = buildTagSuffix(ctx)
	}

	tag, err := latestTag(ctx.GitClient, tagSuffix)
	if err != nil {
		return err
	}

	if tag == "" {
		log.Debug("repository not tagged with version")
	} else {
		log.WithField("version", tag).Debug("identified latest version within repository")
	}

	ctx.CurrentVersion, _ = semver.Parse(tag)

	glog, err := ctx.GitClient.Log(git.WithRefRange(git.HeadRef, tag))
	if err != nil {
		return err
	}

	// Configure parse options based on context
	parseOptions := semver.ParseOptions{
		TrimHeader:     ctx.Changelog.TrimHeader,
		PreReleaseMode: ctx.PreReleaseMode,
		PatchTypes:     ctx.PatchTypes,
	}

	// Identify any commit that will trigger the largest semantic version bump
	inc := semver.ParseLogWithOptions(glog.Commits, parseOptions)

	if inc == semver.NoIncrement {
		ctx.NoVersionChanged = true
		log.Warn("no commits trigger a change in semantic version")
		return nil
	}

	log.WithField("increment", string(inc)).Info("largest increment detected from commits")

	if tag == "" {
		tag = "v0.0.0"
	}

	// Remove the prefix if needed
	if ctx.NoPrefix {
		tag = strings.TrimPrefix(tag, "v")
	}

	pver, _ := semv.NewVersion(tag)
	nxt := *pver

	if ctx.IgnoreExistingPrerelease {
		log.Info("stripped existing prerelease metadata from version")
		nxt, _ = nxt.SetPrerelease("")
		nxt, _ = nxt.SetMetadata("")
	}

	// Apply version increment based on type
	nxt, err = t.applyVersionIncrement(ctx, nxt, inc)
	if err != nil {
		return err
	}

	ctx.NextVersion = semver.Version{
		Prefix:     ctx.CurrentVersion.Prefix,
		Patch:      nxt.Patch(),
		Minor:      nxt.Minor(),
		Major:      nxt.Major(),
		Prerelease: nxt.Prerelease(),
		Metadata:   nxt.Metadata(),
		Raw:        nxt.Original(),
	}

	log.WithField("version", ctx.NextVersion.Raw).Info("identified next semantic version")
	return nil
}

// applyVersionIncrement applies the version increment and handles pre-release logic
func (t Task) applyVersionIncrement(ctx *context.Context, nxt semv.Version, inc semver.Increment) (semv.Version, error) {
	// Apply base version increment
	nxt = t.applyBaseIncrement(nxt, inc)

	// Handle pre-release versioning
	if inc == semver.PreReleaseIncrement {
		return t.handlePreReleaseIncrement(ctx, nxt)
	}

	// Handle explicit pre-release settings
	if ctx.Prerelease != "" {
		nxt, _ = nxt.SetPrerelease(ctx.Prerelease)
		log.WithField("prerelease", ctx.Prerelease).Debug("applied explicit prerelease")
	}

	// Handle metadata
	if ctx.Metadata != "" {
		nxt, _ = nxt.SetMetadata(ctx.Metadata)
		log.WithField("metadata", ctx.Metadata).Debug("applied metadata")
	}

	return nxt, nil
}

// applyBaseIncrement applies major/minor/patch increment
func (t Task) applyBaseIncrement(nxt semv.Version, inc semver.Increment) semv.Version {
	switch inc {
	case semver.MajorIncrement:
		return nxt.IncMajor()
	case semver.MinorIncrement:
		return nxt.IncMinor()
	case semver.PatchIncrement:
		return nxt.IncPatch()
	case semver.PreReleaseIncrement:
		// For pre-release increment, we don't change the base version
		// unless it's a completely new version
		if nxt.Major() == 0 && nxt.Minor() == 0 && nxt.Patch() == 0 {
			return nxt.IncPatch()
		}
		return nxt
	default:
		return nxt
	}
}

// handlePreReleaseIncrement handles pre-release version increment logic
func (t Task) handlePreReleaseIncrement(ctx *context.Context, nxt semv.Version) (semv.Version, error) {
	var err error

	// Determine pre-release identifier
	preReleaseType := t.determinePreReleaseType(ctx)

	if ctx.PreReleaseMode == semver.PreReleaseModeAuto || ctx.PreReleaseMode == semver.PreReleaseModeAlways {
		nxt, err = t.autoIncrementPreRelease(nxt, ctx.CurrentVersion.Prerelease, preReleaseType)
		if err != nil {
			return nxt, err
		}

		log.WithFields(log.Fields{
			"auto_increment": true,
			"prerelease":     nxt.Prerelease(),
			"type":           preReleaseType,
		}).Debug("auto-incremented pre-release version")
	} else if ctx.Prerelease != "" {
		// Use explicit pre-release setting
		nxt, err = nxt.SetPrerelease(ctx.Prerelease)
		if err != nil {
			return nxt, err
		}

		log.WithField("prerelease", ctx.Prerelease).Debug("applied explicit prerelease")
	} else {
		// Create new pre-release version with default type
		newPreRelease := fmt.Sprintf("%s.1", preReleaseType)
		nxt, err = nxt.SetPrerelease(newPreRelease)
		if err != nil {
			return nxt, err
		}

		log.WithField("prerelease", newPreRelease).Debug("created new pre-release version")
	}

	// Apply metadata if provided
	if ctx.Metadata != "" {
		nxt, err = nxt.SetMetadata(ctx.Metadata)
		if err != nil {
			return nxt, err
		}
	}

	return nxt, nil
}

// determinePreReleaseType determines the appropriate pre-release type
func (t Task) determinePreReleaseType(ctx *context.Context) string {
	// If explicit pre-release is provided, extract the type
	if ctx.Prerelease != "" {
		preType, _ := extractPreReleaseTypeAndNumber(ctx.Prerelease)
		if preType != "" {
			return preType
		}
	}

	// Default to alpha for new pre-releases
	return "alpha"
}

// autoIncrementPreRelease handles automatic pre-release version increment
func (t Task) autoIncrementPreRelease(nxt semv.Version, currentPreRelease, targetType string) (semv.Version, error) {
	if currentPreRelease == "" {
		// No existing pre-release, create new one
		return nxt.SetPrerelease(fmt.Sprintf("%s.1", targetType))
	}

	// Parse current pre-release
	currentType, currentNumber := extractPreReleaseTypeAndNumber(currentPreRelease)

	// Determine next version based on pre-release evolution rules
	nextType, nextNumber := t.evolvePreReleaseVersion(currentType, currentNumber, targetType)

	newPreRelease := fmt.Sprintf("%s.%d", nextType, nextNumber)
	return nxt.SetPrerelease(newPreRelease)
}

// evolvePreReleaseVersion implements industry-standard pre-release evolution rules
func (t Task) evolvePreReleaseVersion(currentType string, currentNumber int, targetType string) (string, int) {
	// Define pre-release type hierarchy: alpha -> beta -> rc -> release
	hierarchy := map[string]int{
		"alpha": 1,
		"beta":  2,
		"rc":    3,
	}

	currentLevel := hierarchy[currentType]
	targetLevel := hierarchy[targetType]

	// If target type is not in hierarchy, use it directly
	if targetLevel == 0 {
		if currentType == targetType {
			return targetType, currentNumber + 1
		}
		return targetType, 1
	}

	// If current type is not in hierarchy, start fresh with target
	if currentLevel == 0 {
		return targetType, 1
	}

	// Apply hierarchy rules
	if targetLevel > currentLevel {
		// Moving to a higher level (e.g., alpha -> beta)
		return targetType, 1
	} else if targetLevel == currentLevel {
		// Same level, increment number
		return targetType, currentNumber + 1
	} else {
		// Moving to a lower level (unusual, but handle it)
		return targetType, 1
	}
}

// extractPreReleaseTypeAndNumber extracts the type and numeric suffix from a pre-release string
func extractPreReleaseTypeAndNumber(prerelease string) (string, int) {
	if prerelease == "" {
		return "", 0
	}

	// Define regex patterns for different pre-release formats
	patterns := []struct {
		regex *regexp.Regexp
		desc  string
	}{
		{regexp.MustCompile(`^([a-zA-Z]+)\.(\d+)$`), "type.number"}, // alpha.1, beta.2
		{regexp.MustCompile(`^([a-zA-Z]+)-(\d+)$`), "type-number"},  // alpha-1, beta-2
		{regexp.MustCompile(`^([a-zA-Z]+)(\d+)$`), "typenumber"},    // alpha1, beta2
		{regexp.MustCompile(`^([a-zA-Z]+)$`), "type only"},          // alpha, beta
	}

	for _, pattern := range patterns {
		matches := pattern.regex.FindStringSubmatch(prerelease)

		if len(matches) >= 2 {
			preType := strings.ToLower(matches[1])
			number := 1 // default number

			if len(matches) > 2 && matches[2] != "" {
				if num, err := strconv.ParseUint(matches[2], 10, 32); err == nil {
					number = int(num)
				}
			}

			return preType, number
		}
	}

	// Fallback: treat the whole string as type with number 1
	return strings.ToLower(prerelease), 1
}

// Utility functions

func latestTag(gitc *git.Client, suffix string) (string, error) {
	tags, err := gitc.Tags(git.WithShellGlob("*.*.*"),
		git.WithSortBy(git.CreatorDateDesc, git.VersionDesc))
	if err != nil {
		return "", err
	}

	if len(tags) == 0 {
		return "", nil
	}

	if suffix == "" {
		return tags[0], nil
	}

	for _, tag := range tags {
		if strings.HasSuffix(tag, suffix) {
			return tag, nil
		}
	}

	return "", nil
}

func buildTagSuffix(ctx *context.Context) string {
	var suffix string
	if ctx.Prerelease != "" {
		suffix = fmt.Sprintf("-%s", ctx.Prerelease)
		if ctx.Metadata != "" {
			suffix = fmt.Sprintf("%s+%s", suffix, ctx.Metadata)
		}
	}
	return suffix
}
