package semver

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	git "github.com/purpleclay/gitz"
)

const (
	// PreReleaseModeNone means no automatic pre-release handling (default)
	PreReleaseModeNone string = ""
	// PreReleaseModeAuto detects pre-release based on commit message tags
	PreReleaseModeAuto string = "auto"
	// PreReleaseModeAlways forces all version increments to be pre-release
	PreReleaseModeAlways string = "always"
)

// Increment defines the different types of increment that can be performed
// against a semantic version
type Increment string

// ParseOptions contains configuration for parsing commit messages
type ParseOptions struct {
	TrimHeader     bool
	PreReleaseMode string   // Enable automatic pre-release increment
	PatchTypes     []string // Custom patch types (e.g., "fix", "perf", "security")
	PreReleaseType string   // Default pre-release type (alpha, beta, rc)
}

const (
	// NoIncrement represents no increment change to a semantic version
	NoIncrement Increment = "None"
	// PatchIncrement represents a patch increment (1.0.x) to a semantic version
	PatchIncrement Increment = "Patch"
	// MinorIncrement represents a minor increment (1.x.0) to a semantic version
	MinorIncrement Increment = "Minor"
	// MajorIncrement represents a major increment (x.0.0) to a semantic version
	MajorIncrement Increment = "Major"
	// PreReleaseIncrement represents a pre-release increment (1.0.0-alpha.x)
	PreReleaseIncrement Increment = "PreRelease"
)

const (
	colonSpace     = ": "
	featUpper      = "FEAT"
	breaking       = "BREAKING CHANGE: "
	breakingHyphen = "BREAKING-CHANGE: "
	breakingBang   = '!'
)

// Default patch types
var defaultPatchTypes = []string{"fix", "perf", "security"}

// Default pre-release types for auto-detection
var defaultPreReleaseTypes = []string{
	"alpha", "beta", "rc", "dev", "canary", "preview", "snapshot",
}

// Pre-release type hierarchy
var preReleaseHierarchy = map[string]int{
	"alpha": 1,
	"beta":  2,
	"rc":    3,
}

// ParseLog will identify the maximum semantic increment by parsing the commit
// log against the conventional commit standards
func ParseLog(log []git.LogEntry) Increment {
	return ParseLogWithOptions(log, ParseOptions{
		TrimHeader:     false,
		PreReleaseMode: PreReleaseModeNone,
		PatchTypes:     defaultPatchTypes,
		PreReleaseType: "alpha",
	})
}

// ParseLogWithOptions parses commit log with custom options
func ParseLogWithOptions(log []git.LogEntry, options ParseOptions) Increment {
	hasPreReleaseIndicators := false

	maxIncrement := NoIncrement

	patchTypes := mergeAndDeduplicatePatchTypes(options)

	for _, entry := range log {
		// Extract conventional commit type
		commitType, hasValidType := extractCommitType(entry.Message, options.TrimHeader)
		if !hasValidType {
			continue
		}

		// Check for pre-release context in commits
		if options.PreReleaseMode == PreReleaseModeAuto && hasPreReleaseContext(commitType, entry.Message) {
			hasPreReleaseIndicators = true
		}

		// Determine semantic increment based on commit type
		increment := determineIncrementType(commitType, entry.Message, patchTypes)
		maxIncrement = updateMaxIncrement(maxIncrement, increment)
	}

	// Convert to pre-release increment if we found pre-release indicators
	if hasPreReleaseIndicators || options.PreReleaseMode == PreReleaseModeAlways {
		return PreReleaseIncrement // Default pre-release without base version change
	}

	return maxIncrement
}

// extractCommitType extracts and validates the conventional commit type
func extractCommitType(message string, trimHeader bool) (string, bool) {
	colonSpaceIdx := strings.Index(message, colonSpace)
	if colonSpaceIdx == -1 {
		return "", false
	}

	startIdx := 0
	if trimHeader {
		startIdx = FindStartIdx(message)
	}

	if startIdx >= colonSpaceIdx {
		return "", false
	}

	commitType := message[startIdx:colonSpaceIdx]
	return commitType, true
}

// FindStartIdx finds the starting index of the conventional commit type
func FindStartIdx(msg string) int {
	colonIdx := strings.Index(msg, colonSpace)
	if colonIdx == -1 {
		return 0
	}

	trimmedMsg := msg[:colonIdx]
	leadingLineBreakIdx := strings.LastIndex(trimmedMsg, "\n")
	if leadingLineBreakIdx == -1 {
		return 0
	}

	return leadingLineBreakIdx + 1
}

// isBreakingChange checks if the commit represents a breaking change
func isBreakingChange(commitType, fullMessage string) bool {
	// Check for breaking change indicator in commit type (e.g., feat!)
	if strings.HasSuffix(commitType, "!") {
		return true
	}

	// Check for breaking change in commit message body/footer
	return strings.Contains(fullMessage, breaking) ||
		strings.Contains(fullMessage, breakingHyphen)
}

// hasPreReleaseContext checks if the commit has pre-release context
func hasPreReleaseContext(commitType, message string) bool {
	// Check commit type scope for pre-release indicators
	if hasPreReleaseInScope(commitType) {
		return true
	}

	// Check message content for pre-release keywords
	return hasPreReleaseInMessage(message)
}

// hasPreReleaseInScope checks if the commit type scope contains pre-release indicators
func hasPreReleaseInScope(commitType string) bool {
	// Extract scope from commit type: feat(scope) -> scope
	re := regexp.MustCompile(`\w+$([^)]+)$`)
	matches := re.FindStringSubmatch(commitType)
	if len(matches) < 2 {
		return false
	}

	scope := strings.ToLower(matches[1])
	for _, preType := range defaultPreReleaseTypes {
		if scope == preType || strings.Contains(scope, preType) {
			return true
		}
	}

	return false
}

// hasPreReleaseInMessage checks if the commit message contains pre-release keywords
func hasPreReleaseInMessage(message string) bool {
	messageLower := strings.ToLower(message)

	// Look for explicit pre-release mentions
	preReleaseKeywords := []string{
		"pre-release", "prerelease", "alpha release", "beta release",
		"rc release", "preview release", "experimental", "unstable",
		"dev build", "canary", "snapshot",
	}

	for _, keyword := range preReleaseKeywords {
		if strings.Contains(messageLower, keyword) {
			return true
		}
	}

	// Look for version pattern mentions (e.g., "alpha.1", "beta.2")
	versionPattern := regexp.MustCompile(`(alpha|beta|rc|dev|canary|preview|snapshot)\.?\d*`)
	return versionPattern.MatchString(messageLower)
}

// determineIncrementType determines the increment type based on conventional commit type
func determineIncrementType(commitType string, message string, patchTypes []string) Increment {
	if isBreakingChange(commitType, message) {
		return MajorIncrement
	}

	// Clean commit type (remove scope and breaking indicator)
	cleanType := cleanCommitType(commitType)
	cleanTypeUpper := strings.ToUpper(cleanType)

	// Check for feature commits (minor increment)
	if cleanTypeUpper == featUpper {
		return MinorIncrement
	}

	// Check for patch type commits
	for _, patchType := range patchTypes {
		if strings.ToUpper(patchType) == cleanTypeUpper {
			return PatchIncrement
		}
	}

	return NoIncrement
}

// cleanCommitType removes scope and breaking indicators from commit type
func cleanCommitType(commitType string) string {
	// Remove breaking indicator (!)
	cleaned := strings.TrimSuffix(commitType, "!")

	// Remove scope: feat(scope) -> feat
	if idx := strings.Index(cleaned, "("); idx != -1 {
		cleaned = cleaned[:idx]
	}

	return cleaned
}

// updateMaxIncrement returns the higher priority increment
func updateMaxIncrement(current, new Increment) Increment {
	priority := map[Increment]int{
		NoIncrement:         0,
		PatchIncrement:      1,
		MinorIncrement:      2,
		MajorIncrement:      3,
		PreReleaseIncrement: 4,
	}

	if priority[new] > priority[current] {
		return new
	}
	return current
}

// GetIncrementDescription returns a human-readable description of the increment
func GetIncrementDescription(increment Increment) string {
	descriptions := map[Increment]string{
		NoIncrement:         "No version change",
		PatchIncrement:      "Patch version increment (bug fixes, performance improvements)",
		MinorIncrement:      "Minor version increment (new features)",
		MajorIncrement:      "Major version increment (breaking changes)",
		PreReleaseIncrement: "Pre-release version increment",
	}

	if desc, ok := descriptions[increment]; ok {
		return desc
	}
	return "Unknown increment type"
}

// IsPreRelease returns true if the increment represents a pre-release
func IsPreRelease(increment Increment) bool {
	return increment == PreReleaseIncrement
}

// PreReleaseManager handles pre-release version increment logic
type PreReleaseManager struct {
	currentVersion string
	defaultType    string
}

// NewPreReleaseManager creates a new pre-release manager
func NewPreReleaseManager(currentVersion string, defaultType string) *PreReleaseManager {
	if defaultType == "" {
		defaultType = "alpha"
	}
	return &PreReleaseManager{
		currentVersion: currentVersion,
		defaultType:    defaultType,
	}
}

// createPreReleaseWithIncrement creates pre-release version with base increment
func (pm *PreReleaseManager) createPreReleaseWithIncrement(version Version, baseIncrement Increment, targetType string) string {
	newVersion := pm.applyBaseIncrement(version, baseIncrement)
	newVersion.Prerelease = fmt.Sprintf("%s.1", targetType)
	newVersion.Metadata = "" // Clear metadata on version increment
	return newVersion.String()
}

// applyBaseIncrement applies major/minor/patch increment to base version
func (pm *PreReleaseManager) applyBaseIncrement(version Version, increment Increment) Version {
	newVersion := version

	switch increment {
	case MajorIncrement:
		newVersion.Major++
		newVersion.Minor = 0
		newVersion.Patch = 0
		newVersion.Prerelease = ""
		newVersion.Metadata = ""
	case MinorIncrement:
		newVersion.Minor++
		newVersion.Patch = 0
		newVersion.Prerelease = ""
		newVersion.Metadata = ""
	case PatchIncrement:
		newVersion.Patch++
		newVersion.Prerelease = ""
		newVersion.Metadata = ""
	}

	return newVersion
}

// incrementExistingPreRelease increments an existing pre-release version
func (pm *PreReleaseManager) incrementExistingPreRelease(version Version, targetType string) (string, error) {
	if version.Prerelease == "" {
		// No existing pre-release, create new one
		version.Prerelease = fmt.Sprintf("%s.1", targetType)
		return version.String(), nil
	}

	// Parse and evolve existing pre-release
	currentType, currentNumber, err := pm.parsePreReleaseInfo(version.Prerelease)
	if err != nil {
		return "", err
	}

	newType, newNumber := pm.evolvePreRelease(currentType, currentNumber, targetType)
	version.Prerelease = fmt.Sprintf("%s.%d", newType, newNumber)

	return version.String(), nil
}

// parsePreReleaseInfo parses pre-release string to extract type and number
func (pm *PreReleaseManager) parsePreReleaseInfo(prerelease string) (string, int, error) {
	if prerelease == "" {
		return "", 0, fmt.Errorf("empty pre-release string")
	}

	// Support multiple formats: alpha.1, alpha-1, alpha1, alpha
	patterns := []struct {
		regex *regexp.Regexp
		desc  string
	}{
		{regexp.MustCompile(`^([a-zA-Z]+)\.(\d+)$`), "type.number"},
		{regexp.MustCompile(`^([a-zA-Z]+)-(\d+)$`), "type-number"},
		{regexp.MustCompile(`^([a-zA-Z]+)(\d+)$`), "typenumber"},
		{regexp.MustCompile(`^([a-zA-Z]+)$`), "type only"},
	}

	for _, pattern := range patterns {
		matches := pattern.regex.FindStringSubmatch(prerelease)

		if len(matches) >= 2 {
			preType := strings.ToLower(matches[1])
			number := 1 // default number

			if len(matches) > 2 && matches[2] != "" {
				if num, err := strconv.Atoi(matches[2]); err == nil {
					number = num
				}
			}

			return preType, number, nil
		}
	}

	// Fallback: treat the whole string as type with number 1
	return strings.ToLower(prerelease), 1, nil
}

// evolvePreRelease handles pre-release version evolution
func (pm *PreReleaseManager) evolvePreRelease(currentType string, currentNumber int, targetType string) (string, int) {
	currentLevel := preReleaseHierarchy[currentType]
	targetLevel := preReleaseHierarchy[targetType]

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

// ParseVersion parses a semantic version string
func ParseVersion(versionStr string) (Version, error) {
	var version Version

	// Remove 'v' prefix if present
	versionStr = strings.TrimPrefix(versionStr, "v")

	// Split by '+' to separate metadata
	parts := strings.Split(versionStr, "+")
	if len(parts) > 1 {
		version.Metadata = parts[1]
	}
	versionStr = parts[0]

	// Split by '-' to separate pre-release
	parts = strings.Split(versionStr, "-")
	if len(parts) > 1 {
		version.Prerelease = strings.Join(parts[1:], "-")
	}
	versionStr = parts[0]

	// Parse major.minor.patch
	versionParts := strings.Split(versionStr, ".")
	if len(versionParts) != 3 {
		return version, fmt.Errorf("invalid version format: %s", versionStr)
	}

	var err error
	if version.Major, err = strconv.ParseInt(versionParts[0], 10, 64); err != nil {
		return version, fmt.Errorf("invalid major version: %s", versionParts[0])
	}
	if version.Minor, err = strconv.ParseInt(versionParts[1], 10, 64); err != nil {
		return version, fmt.Errorf("invalid minor version: %s", versionParts[1])
	}
	if version.Patch, err = strconv.ParseInt(versionParts[2], 10, 64); err != nil {
		return version, fmt.Errorf("invalid patch version: %s", versionParts[2])
	}

	return version, nil
}

// 内联实现版本
func mergeAndDeduplicatePatchTypes(options ParseOptions) []string {
	typeMap := make(map[string]string)
	var result []string

	// 合并默认类型和自定义类型
	allTypes := append(defaultPatchTypes, options.PatchTypes...)

	for _, pType := range allTypes {
		key := strings.ToLower(strings.TrimSpace(pType))
		if key != "" && typeMap[key] == "" {
			typeMap[key] = pType
			result = append(result, pType)
		}
	}

	return result
}
