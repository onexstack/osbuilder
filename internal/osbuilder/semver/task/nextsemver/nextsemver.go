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
	log.WithField("increment", inc).Info("Detected increment")

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
	// var err error

	// 判断是否处于预发布模式
	isPreReleaseMode := t.isPreReleaseMode(ctx, inc)

	if isPreReleaseMode {
		// 预发布模式处理
		return t.handlePreReleaseMode(ctx, nxt, inc)
	} else {
		// 正式版本模式处理
		return t.handleReleaseMode(ctx, nxt, inc)
	}
}

// isPreReleaseMode 判断是否处于预发布模式
func (t Task) isPreReleaseMode(ctx *context.Context, inc semver.Increment) bool {
	// 如果increment本身就是预发布类型
	if inc == semver.PreReleaseIncrement {
		return true
	}

	// 如果明确指定了预发布标识
	if ctx.Prerelease != "" {
		return true
	}

	// 如果当前版本是预发布版本且在预发布周期内
	if ctx.CurrentVersion.Prerelease != "" && ctx.PreReleaseMode != semver.PreReleaseModeNone {
		return true
	}

	return false
}

// handleReleaseMode 处理正式版本模式
func (t Task) handleReleaseMode(ctx *context.Context, nxt semv.Version, inc semver.Increment) (semv.Version, error) {
	var err error

	// 应用基础版本递增
	nxt = t.applyBaseIncrement(nxt, inc)

	// 清除预发布信息（正式版本不应该有预发布标识）
	nxt, _ = nxt.SetPrerelease("")

	// 应用元数据（如果有）
	if ctx.Metadata != "" {
		nxt, err = nxt.SetMetadata(ctx.Metadata)
		if err != nil {
			return nxt, fmt.Errorf("failed to set metadata: %w", err)
		}
		log.WithField("metadata", ctx.Metadata).Debug("applied metadata to release version")
	}

	return nxt, nil
}

// handlePreReleaseMode 处理预发布模式
func (t Task) handlePreReleaseMode(ctx *context.Context, nxt semv.Version, inc semver.Increment) (semv.Version, error) {
	// var err error

	currentPreRelease := ctx.CurrentVersion.Prerelease

	// 确定目标版本号
	targetVersion := t.determineTargetVersion(ctx.CurrentVersion, inc)

	if currentPreRelease == "" {
		// 场景1: 从正式版本开始预发布周期
		return t.startNewPreReleaseFromRelease(ctx, targetVersion.IncPatch(), inc)
	} else {
		// 场景2: 在预发布周期内
		return t.continuePreRelease(ctx, nxt, targetVersion, inc)
	}
}

// determineTargetVersion 根据当前版本和增量类型确定目标版本号
func (t Task) determineTargetVersion(currentVersion semver.Version, inc semver.Increment) semv.Version {
	// 构建当前版本的semver对象
	currentSemver := fmt.Sprintf("v%d.%d.%d", currentVersion.Major, currentVersion.Minor, currentVersion.Patch)
	if currentVersion.Prerelease != "" {
		currentSemver += "-" + currentVersion.Prerelease
	}

	current, _ := semv.NewVersion(currentSemver)
	target := *current

	// 如果当前是预发布版本，先清除预发布标识获得基础版本
	if currentVersion.Prerelease != "" {
		target, _ = target.SetPrerelease("")
	}

	// 根据增量类型确定目标基础版本
	switch inc {
	case semver.MajorIncrement:
		target = target.IncMajor()
	case semver.MinorIncrement:
		target = target.IncMinor()
	case semver.PatchIncrement:
		target = target.IncPatch()
	case semver.PreReleaseIncrement:
		// 纯预发布增量，不改变基础版本号
		// 保持当前的基础版本
	}

	return target
}

// startNewPreReleaseFromRelease 从正式版本开始新的预发布周期
func (t Task) startNewPreReleaseFromRelease(ctx *context.Context, targetVersion semv.Version, inc semver.Increment) (semv.Version, error) {
	var err error

	log.Info("Starting new pre-release cycle from release version")
	log.WithField("targetVersion", targetVersion.String()).Info("Target base version")

	// 确定预发布类型
	preReleaseType := t.determinePreReleaseType(ctx)

	// 创建新的预发布版本: 目标版本-type.1
	newPreRelease := fmt.Sprintf("%s.1", preReleaseType)
	targetVersion, err = targetVersion.SetPrerelease(newPreRelease)
	if err != nil {
		return targetVersion, fmt.Errorf("failed to set prerelease: %w", err)
	}

	log.WithFields(log.Fields{
		"from":       ctx.CurrentVersion.Raw,
		"to":         targetVersion.String(),
		"prerelease": newPreRelease,
		"type":       preReleaseType,
	}).Info("started new pre-release cycle")

	// 应用元数据
	if ctx.Metadata != "" {
		targetVersion, err = targetVersion.SetMetadata(ctx.Metadata)
		if err != nil {
			return targetVersion, fmt.Errorf("failed to set metadata: %w", err)
		}
	}

	return targetVersion, nil
}

// continuePreRelease 在预发布周期内继续
func (t Task) continuePreRelease(ctx *context.Context, nxt semv.Version, targetVersion semv.Version, inc semver.Increment) (semv.Version, error) {
	// var err error

	currentVersion := ctx.CurrentVersion

	// 比较当前基础版本和目标基础版本
	currentBase := fmt.Sprintf("%d.%d.%d", currentVersion.Major, currentVersion.Minor, currentVersion.Patch)
	targetBase := fmt.Sprintf("%d.%d.%d", targetVersion.Major(), targetVersion.Minor(), targetVersion.Patch())

	log.WithFields(log.Fields{"currentBase": currentBase, "targetBase": targetBase}).Info("Continue pre-release")

	if currentBase != targetBase {
		fmt.Println("666666666666666666666666666666666666-1")
		// 场景2a: 基础版本变化，开始新的预发布周期
		return t.startNewPreReleaseFromPreRelease(ctx, targetVersion, inc)
	} else {
		// 场景2b: 基础版本相同，递增预发布版本
		return t.incrementPreReleaseVersion(ctx, nxt, inc)
	}
}

// startNewPreReleaseFromPreRelease 从预发布版本开始新的预发布周期（基础版本发生变化）
func (t Task) startNewPreReleaseFromPreRelease(ctx *context.Context, targetVersion semv.Version, inc semver.Increment) (semv.Version, error) {
	var err error

	log.Info("Starting new pre-release cycle due to base version change")

	// 确定预发布类型
	preReleaseType := t.determinePreReleaseType(ctx)

	// 重置预发布版本: 新基础版本-type.1
	newPreRelease := fmt.Sprintf("%s.1", preReleaseType)
	targetVersion, err = targetVersion.SetPrerelease(newPreRelease)
	if err != nil {
		return targetVersion, fmt.Errorf("failed to set prerelease: %w", err)
	}

	log.WithFields(log.Fields{
		"from":           ctx.CurrentVersion.Raw,
		"to":             targetVersion.String(),
		"base_changed":   true,
		"new_prerelease": newPreRelease,
	}).Info("started new pre-release cycle due to base version change")

	// 应用元数据
	if ctx.Metadata != "" {
		targetVersion, err = targetVersion.SetMetadata(ctx.Metadata)
		if err != nil {
			return targetVersion, fmt.Errorf("failed to set metadata: %w", err)
		}
	}

	return targetVersion, nil
}

// incrementPreReleaseVersion 递增预发布版本号
func (t Task) incrementPreReleaseVersion(ctx *context.Context, nxt semv.Version, inc semver.Increment) (semv.Version, error) {
	var err error

	log.Info("Incrementing pre-release version")

	currentPreRelease := ctx.CurrentVersion.Prerelease
	currentType, currentNumber := extractPreReleaseTypeAndNumber(currentPreRelease)

	// 确定目标预发布类型
	targetType := t.determinePreReleaseType(ctx)

	// 计算下一个预发布版本
	nextType, nextNumber := t.evolvePreReleaseVersion(currentType, currentNumber, targetType)
	newPreRelease := fmt.Sprintf("%s.%d", nextType, nextNumber)

	// 保持基础版本不变，只更新预发布标识
	result := nxt // nxt应该已经包含了当前的基础版本号
	result, err = result.SetPrerelease(newPreRelease)
	if err != nil {
		return result, fmt.Errorf("failed to set prerelease: %w", err)
	}

	log.WithFields(log.Fields{
		"from":         currentPreRelease,
		"to":           newPreRelease,
		"current_type": currentType,
		"target_type":  targetType,
		"number":       nextNumber,
	}).Info("incremented pre-release version")

	// 应用元数据
	if ctx.Metadata != "" {
		result, err = result.SetMetadata(ctx.Metadata)
		if err != nil {
			return result, fmt.Errorf("failed to set metadata: %w", err)
		}
	}

	return result, nil
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
	default:
		return nxt
	}
}

// determinePreReleaseType determines the appropriate pre-release type
func (t Task) determinePreReleaseType(ctx *context.Context) string {
	// 如果明确指定了预发布标识，提取类型
	if ctx.Prerelease != "" {
		preType, _ := extractPreReleaseTypeAndNumber(ctx.Prerelease)
		if preType != "" {
			return preType
		}
		// 如果解析失败，直接使用原值
		return strings.ToLower(ctx.Prerelease)
	}

	// 如果当前版本有预发布标识，继续使用相同类型
	if ctx.CurrentVersion.Prerelease != "" {
		preType, _ := extractPreReleaseTypeAndNumber(ctx.CurrentVersion.Prerelease)
		if preType != "" {
			return preType
		}
	}

	// 默认使用 alpha
	return "alpha"
}

// evolvePreReleaseVersion implements industry-standard pre-release evolution rules
func (t Task) evolvePreReleaseVersion(currentType string, currentNumber int, targetType string) (string, int) {
	// Define pre-release type hierarchy: alpha -> beta -> rc
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
