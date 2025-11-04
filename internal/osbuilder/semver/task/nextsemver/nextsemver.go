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

	fmt.Println("11111111111111111111111-1", tag, tagSuffix)
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
	fmt.Println("111111111111111111111111111111111111111111111121233333333333333333333333333-44", inc)

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
	var err error

	// Apply base version increment first
	nxt = t.applyBaseIncrement(nxt, inc)
	fmt.Println("6666666666666666666666666666666666666666666666666666-1", inc)

	// Handle different increment types
	switch inc {
	case semver.PreReleaseIncrement,
		semver.PreReleasePatchIncrement,
		semver.PreReleaseMinorIncrement,
		semver.PreReleaseMajorIncrement:
		// Handle pre-release increment logic
		return t.handlePreReleaseIncrement(ctx, nxt)
	default:
		// For regular increments (Major, Minor, Patch, NoIncrement)
		// Check if we need to add pre-release or metadata

		// Handle explicit pre-release settings
		if ctx.Prerelease != "" {
			nxt, err = nxt.SetPrerelease(ctx.Prerelease)
			if err != nil {
				return nxt, fmt.Errorf("failed to set prerelease: %w", err)
			}
			log.WithField("prerelease", ctx.Prerelease).Debug("applied explicit prerelease")
		}
	}

	// Handle metadata (applies to all increment types)
	if ctx.Metadata != "" {
		nxt, err = nxt.SetMetadata(ctx.Metadata)
		if err != nil {
			return nxt, fmt.Errorf("failed to set metadata: %w", err)
		}
		log.WithField("metadata", ctx.Metadata).Debug("applied metadata")
	}

	return nxt, nil
}

// applyBaseIncrement applies major/minor/patch increment
func (t Task) applyBaseIncrement(nxt semv.Version, inc semver.Increment) semv.Version {
	switch inc {
	case semver.MajorIncrement, semver.PreReleaseMajorIncrement:
		return nxt.IncMajor()
	case semver.MinorIncrement, semver.PreReleaseMinorIncrement:
		return nxt.IncMinor()
	case semver.PatchIncrement, semver.PreReleasePatchIncrement:
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
	fmt.Println("555555555555555555555555555555555555555", nxt.String(), "#", ctx.CurrentVersion.Prerelease, "#", preReleaseType)

	if ctx.PreReleaseMode == semver.PreReleaseModeAuto || ctx.PreReleaseMode == semver.PreReleaseModeAlways {
		nxt, err = t.autoIncrementPreRelease(nxt, ctx.CurrentVersion, preReleaseType)
		if err != nil {
			return nxt, err
		}
		fmt.Println("111111111111111111111111111111111", ctx.CurrentVersion.Prerelease, "2222", preReleaseType, nxt.String())

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
func (t Task) autoIncrementPreRelease(nxt semv.Version, cur semver.Version, targetType string) (semv.Version, error) {
	currentPreRelease := cur.Prerelease
	// 场景1：当前是正式版本，需要创建下一个预发布版本
	if currentPreRelease == "" {
		// 从正式版本创建预发布版本，需要根据增量类型确定基础版本
		// 例如：v0.1.3 -> v0.1.4-alpha.1 (patch)
		//      v0.1.3 -> v0.2.0-alpha.1 (minor)
		//      v0.1.3 -> v1.0.0-alpha.1 (major)

		// nxt 已经包含了正确的增量版本号，只需添加预发布标识
		newPreRelease := fmt.Sprintf("%s.1", targetType)
		return nxt.SetPrerelease(newPreRelease)
	}

	// Parse current pre-release
	currentType, currentNumber := extractPreReleaseTypeAndNumber(currentPreRelease)

	// 检查当前预发布版本的基础版本号是否与目标版本号匹配
	currentBase, _ := semv.NewVersion(fmt.Sprintf("%d.%d.%d", cur.Major, cur.Minor, cur.Patch))
	nextBase, _ := semv.NewVersion(fmt.Sprintf("%d.%d.%d", nxt.Major(), nxt.Minor(), nxt.Patch()))

	// 如果基础版本号不同，说明有新的增量，重置预发布版本
	if !currentBase.Equal(nextBase) {
		// 基础版本发生变化，创建新的预发布版本
		// 例如：v0.1.3-alpha.2 + minor增量 -> v0.2.0-alpha.1
		newPreRelease := fmt.Sprintf("%s.1", targetType)
		return nxt.SetPrerelease(newPreRelease)
	}

	// 基础版本相同，按照预发布版本演进规则处理
	nextType, nextNumber := t.evolvePreReleaseVersion(currentType, currentNumber, targetType)
	newPreRelease := fmt.Sprintf("%s.%d", nextType, nextNumber)

	// 保持相同的基础版本号，只更新预发布标识
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

	fmt.Println("2222222222222222222222222222132", tags)
	if suffix == "" {
		return tags[0], nil
	}

	fmt.Println("2222222222222222222222222222132-3")
	for _, tag := range tags {
		fmt.Println("2222222222222222222222222222132-3-1", tag, suffix, strings.HasSuffix(tag, suffix))
		if strings.HasSuffix(tag, suffix) {
			fmt.Println("2222222222222222222222222222132-3-2", tag)
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
