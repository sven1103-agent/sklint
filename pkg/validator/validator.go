package validator

import (
	"bytes"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"

	"sklint/internal/parse"
)

const (
	codePathNotFound           = "PATH_NOT_FOUND"
	codePathNotDirectory       = "PATH_NOT_DIRECTORY"
	codeSkillMDMissing         = "SKILL_MD_MISSING"
	codeSkillMDNotFile         = "SKILL_MD_NOT_FILE"
	codeSkillMDSymlink         = "SKILL_MD_SYMLINK"
	codeSkillMDSymlinkInvalid  = "SKILL_MD_SYMLINK_INVALID"
	codeSkillMDSymlinkEscapes  = "SKILL_MD_SYMLINK_ESCAPES_ROOT"
	codeFrontmatterStart       = "FRONTMATTER_START_MISSING"
	codeFrontmatterEnd         = "FRONTMATTER_END_MISSING"
	codeFrontmatterEmpty       = "FRONTMATTER_EMPTY"
	codeFrontmatterInvalidYAML = "FRONTMATTER_INVALID_YAML"
	codeFrontmatterNotMapping  = "FRONTMATTER_NOT_MAPPING"

	codeScriptsNotDir    = "SCRIPTS_NOT_DIRECTORY"
	codeReferencesNotDir = "REFERENCES_NOT_DIRECTORY"
	codeAssetsNotDir     = "ASSETS_NOT_DIRECTORY"

	codeNameMissing             = "NAME_MISSING"
	codeNameNotString           = "NAME_NOT_STRING"
	codeNameTooShort            = "NAME_TOO_SHORT"
	codeNameTooLong             = "NAME_TOO_LONG"
	codeNameInvalidChars        = "NAME_INVALID_CHARS"
	codeNameStartsWithHyphen    = "NAME_STARTS_WITH_HYPHEN"
	codeNameEndsWithHyphen      = "NAME_ENDS_WITH_HYPHEN"
	codeNameConsecutiveHyphens  = "NAME_CONSECUTIVE_HYPHENS"
	codeNameMismatchDirectory   = "NAME_MISMATCH_DIRECTORY"
	codeDescriptionMissing      = "DESCRIPTION_MISSING"
	codeDescriptionNotString    = "DESCRIPTION_NOT_STRING"
	codeDescriptionTooShort     = "DESCRIPTION_TOO_SHORT"
	codeDescriptionTooLong      = "DESCRIPTION_TOO_LONG"
	codeCompatibilityNotString  = "COMPATIBILITY_NOT_STRING"
	codeCompatibilityTooShort   = "COMPATIBILITY_TOO_SHORT"
	codeCompatibilityTooLong    = "COMPATIBILITY_TOO_LONG"
	codeLicenseNotString        = "LICENSE_NOT_STRING"
	codeMetadataNotObject       = "METADATA_NOT_OBJECT"
	codeMetadataValueNotString  = "METADATA_VALUE_NOT_STRING"
	codeAllowedToolsNotString   = "ALLOWED_TOOLS_NOT_STRING"
	codeAllowedToolsEmpty       = "ALLOWED_TOOLS_EMPTY"

	codeSkillMDTooLongLines   = "SKILL_MD_TOO_LONG_LINES"
	codeSkillMDMissingBody    = "SKILL_MD_MISSING_BODY"
	codeUnknownTopLevelKey    = "UNKNOWN_TOP_LEVEL_KEY"
	codeScriptsDirEmpty       = "SCRIPTS_DIR_EMPTY"
	codeReferencesDirEmpty    = "REFERENCES_DIR_EMPTY"
	codeAssetsDirEmpty        = "ASSETS_DIR_EMPTY"
	codeRefContainsDotDot     = "REF_CONTAINS_DOTDOT"
	codeRefTooDeep            = "REF_TOO_DEEP"
	codeRefMissingFile        = "REF_MISSING_FILE"
	codeRefEscapesRoot        = "REF_ESCAPES_ROOT"
)

var (
	namePattern    = regexp.MustCompile(`^[a-z0-9-]+$`)
	linkPattern    = regexp.MustCompile(`!?\[[^\]]*\]\(([^\s)]+)`)
	plainRefPattern = regexp.MustCompile(`(^|\s)(scripts|references|assets)/[^\s)]+`)
)

var knownKeys = map[string]struct{}{
	"name":         {},
	"description":  {},
	"license":      {},
	"compatibility": {},
	"metadata":     {},
	"allowed-tools": {},
}

func ValidateSkill(path string, opts Options) (Result, error) {
	result := Result{}
	absPath, err := filepath.Abs(path)
	if err != nil {
		return result, err
	}
	result.Path = absPath

	info, err := os.Stat(absPath)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			addError(&result, codePathNotFound, fmt.Sprintf("Path '%s' does not exist.", path), "", 0)
			finalizeResult(&result, opts)
			return result, nil
		}
		return result, err
	}
	if !info.IsDir() {
		addError(&result, codePathNotDirectory, fmt.Sprintf("Path '%s' is not a directory.", path), "", 0)
		finalizeResult(&result, opts)
		return result, nil
	}

	checkOptionalDir(absPath, "scripts", codeScriptsNotDir, codeScriptsDirEmpty, &result, opts)
	checkOptionalDir(absPath, "references", codeReferencesNotDir, codeReferencesDirEmpty, &result, opts)
	checkOptionalDir(absPath, "assets", codeAssetsNotDir, codeAssetsDirEmpty, &result, opts)

	skillPath := filepath.Join(absPath, "SKILL.md")
	skillInfo, err := os.Lstat(skillPath)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			addError(&result, codeSkillMDMissing, "SKILL.md is required.", "SKILL.md", 0)
			finalizeResult(&result, opts)
			return result, nil
		}
		return result, err
	}
	if skillInfo.IsDir() {
		addError(&result, codeSkillMDNotFile, "SKILL.md must be a file, not a directory.", "SKILL.md", 0)
		finalizeResult(&result, opts)
		return result, nil
	}

	resolvedSkillPath := skillPath
	if skillInfo.Mode()&os.ModeSymlink != 0 {
		addWarning(&result, opts, codeSkillMDSymlink, "SKILL.md is a symlink.", "SKILL.md", 0)
		resolved, err := filepath.EvalSymlinks(skillPath)
		if err != nil {
			addError(&result, codeSkillMDSymlinkInvalid, "SKILL.md symlink cannot be resolved.", "SKILL.md", 0)
			finalizeResult(&result, opts)
			return result, nil
		}
		if !opts.FollowSymlinks && !isWithinRoot(absPath, resolved) {
			addError(&result, codeSkillMDSymlinkEscapes, "SKILL.md symlink resolves outside the skill directory.", "SKILL.md", 0)
			finalizeResult(&result, opts)
			return result, nil
		}
		resolvedSkillPath = resolved
	}

	content, err := os.ReadFile(resolvedSkillPath)
	if err != nil {
		return result, err
	}

	frontmatter, err := parse.ParseFrontmatter(bytes.NewReader(content))
	if err != nil {
		switch err {
		case parse.ErrFrontmatterStartMissing:
			addError(&result, codeFrontmatterStart, "SKILL.md must begin with '---' frontmatter delimiter.", "SKILL.md", 1)
		case parse.ErrFrontmatterEndMissing:
			addError(&result, codeFrontmatterEnd, "SKILL.md frontmatter must end with '---' delimiter.", "SKILL.md", 0)
		case parse.ErrFrontmatterEmpty:
			addError(&result, codeFrontmatterEmpty, "Frontmatter must contain at least one key.", "SKILL.md", 0)
		default:
			return result, err
		}
		finalizeResult(&result, opts)
		return result, nil
	}

	var node yaml.Node
	if err := yaml.Unmarshal([]byte(frontmatter.YAML), &node); err != nil {
		addError(&result, codeFrontmatterInvalidYAML, fmt.Sprintf("Frontmatter YAML is invalid: %s", err.Error()), "SKILL.md", 0)
		finalizeResult(&result, opts)
		return result, nil
	}

	root, err := mappingRoot(&node)
	if err != nil {
		addError(&result, codeFrontmatterNotMapping, "Frontmatter YAML must be a mapping/object.", "SKILL.md", 0)
		finalizeResult(&result, opts)
		return result, nil
	}
	if len(root.Content) == 0 {
		addError(&result, codeFrontmatterEmpty, "Frontmatter must contain at least one key.", "SKILL.md", 0)
		finalizeResult(&result, opts)
		return result, nil
	}

	keyLines := mapKeyLines(root, frontmatter.YAMLStartLine)

	var data map[string]any
	if err := yaml.Unmarshal([]byte(frontmatter.YAML), &data); err != nil {
		addError(&result, codeFrontmatterInvalidYAML, fmt.Sprintf("Frontmatter YAML is invalid: %s", err.Error()), "SKILL.md", 0)
		finalizeResult(&result, opts)
		return result, nil
	}

	validateName(&result, data, keyLines, filepath.Base(absPath))
	validateDescription(&result, data, keyLines)
	validateCompatibility(&result, data, keyLines)
	validateLicense(&result, data, keyLines)
	validateMetadata(&result, data, keyLines)
	validateAllowedTools(&result, data, keyLines)

	if !opts.NoWarn {
		unknownKeys := collectUnknownKeys(root)
		if len(unknownKeys) > 0 {
			addWarning(&result, opts, codeUnknownTopLevelKey, fmt.Sprintf("Unknown top-level keys: %s", strings.Join(unknownKeys, ", ")), "SKILL.md", 0)
		}
	}

	if frontmatter.LineCount > 500 {
		addWarning(&result, opts, codeSkillMDTooLongLines, fmt.Sprintf("SKILL.md is %d lines; recommended under 500 lines.", frontmatter.LineCount), "SKILL.md", 0)
	}
	if strings.TrimSpace(frontmatter.Body) == "" {
		addWarning(&result, opts, codeSkillMDMissingBody, "SKILL.md body is empty.", "SKILL.md", 0)
	}

	scanReferences(absPath, frontmatter.Body, &result, opts)

	finalizeResult(&result, opts)
	return result, nil
}

func validateName(result *Result, data map[string]any, lines map[string]int, dirName string) {
	value, ok := data["name"]
	if !ok {
		addError(result, codeNameMissing, "Frontmatter 'name' is required.", "SKILL.md", 0)
		return
	}
	name, ok := value.(string)
	if !ok {
		addError(result, codeNameNotString, "Frontmatter 'name' must be a string.", "SKILL.md", lineFor(lines, "name"))
		return
	}
	if len(name) < 1 {
		addError(result, codeNameTooShort, "Frontmatter 'name' must be at least 1 character.", "SKILL.md", lineFor(lines, "name"))
	}
	if len(name) > 64 {
		addError(result, codeNameTooLong, "Frontmatter 'name' must be at most 64 characters.", "SKILL.md", lineFor(lines, "name"))
	}
	if !namePattern.MatchString(name) {
		addError(result, codeNameInvalidChars, "Frontmatter 'name' must use lowercase letters, digits, and hyphens only.", "SKILL.md", lineFor(lines, "name"))
	}
	if strings.HasPrefix(name, "-") {
		addError(result, codeNameStartsWithHyphen, "Frontmatter 'name' must not start with '-'.", "SKILL.md", lineFor(lines, "name"))
	}
	if strings.HasSuffix(name, "-") {
		addError(result, codeNameEndsWithHyphen, "Frontmatter 'name' must not end with '-'.", "SKILL.md", lineFor(lines, "name"))
	}
	if strings.Contains(name, "--") {
		addError(result, codeNameConsecutiveHyphens, "Frontmatter 'name' must not contain consecutive hyphens.", "SKILL.md", lineFor(lines, "name"))
	}
	if name != dirName {
		addError(result, codeNameMismatchDirectory, fmt.Sprintf("Frontmatter name '%s' must match directory name '%s'.", name, dirName), "SKILL.md", lineFor(lines, "name"))
	}
}

func validateDescription(result *Result, data map[string]any, lines map[string]int) {
	value, ok := data["description"]
	if !ok {
		addError(result, codeDescriptionMissing, "Frontmatter 'description' is required.", "SKILL.md", 0)
		return
	}
	desc, ok := value.(string)
	if !ok {
		addError(result, codeDescriptionNotString, "Frontmatter 'description' must be a string.", "SKILL.md", lineFor(lines, "description"))
		return
	}
	if len(desc) < 1 {
		addError(result, codeDescriptionTooShort, "Frontmatter 'description' must be at least 1 character.", "SKILL.md", lineFor(lines, "description"))
	}
	if len(desc) > 1024 {
		addError(result, codeDescriptionTooLong, "Frontmatter 'description' must be at most 1024 characters.", "SKILL.md", lineFor(lines, "description"))
	}
}

func validateCompatibility(result *Result, data map[string]any, lines map[string]int) {
	value, ok := data["compatibility"]
	if !ok {
		return
	}
	comp, ok := value.(string)
	if !ok {
		addError(result, codeCompatibilityNotString, "Frontmatter 'compatibility' must be a string.", "SKILL.md", lineFor(lines, "compatibility"))
		return
	}
	if len(comp) < 1 {
		addError(result, codeCompatibilityTooShort, "Frontmatter 'compatibility' must be at least 1 character.", "SKILL.md", lineFor(lines, "compatibility"))
	}
	if len(comp) > 500 {
		addError(result, codeCompatibilityTooLong, "Frontmatter 'compatibility' must be at most 500 characters.", "SKILL.md", lineFor(lines, "compatibility"))
	}
}

func validateLicense(result *Result, data map[string]any, lines map[string]int) {
	value, ok := data["license"]
	if !ok {
		return
	}
	if _, ok := value.(string); !ok {
		addError(result, codeLicenseNotString, "Frontmatter 'license' must be a string.", "SKILL.md", lineFor(lines, "license"))
	}
}

func validateMetadata(result *Result, data map[string]any, lines map[string]int) {
	value, ok := data["metadata"]
	if !ok {
		return
	}
	switch typed := value.(type) {
	case map[string]any:
		for _, v := range typed {
			if _, ok := v.(string); !ok {
				addError(result, codeMetadataValueNotString, "Frontmatter 'metadata' values must be strings.", "SKILL.md", lineFor(lines, "metadata"))
				return
			}
		}
	case map[any]any:
		for k, v := range typed {
			if _, ok := k.(string); !ok {
				addError(result, codeMetadataNotObject, "Frontmatter 'metadata' must be an object with string keys.", "SKILL.md", lineFor(lines, "metadata"))
				return
			}
			if _, ok := v.(string); !ok {
				addError(result, codeMetadataValueNotString, "Frontmatter 'metadata' values must be strings.", "SKILL.md", lineFor(lines, "metadata"))
				return
			}
		}
	default:
		addError(result, codeMetadataNotObject, "Frontmatter 'metadata' must be an object.", "SKILL.md", lineFor(lines, "metadata"))
	}
}

func validateAllowedTools(result *Result, data map[string]any, lines map[string]int) {
	value, ok := data["allowed-tools"]
	if !ok {
		return
	}
	tools, ok := value.(string)
	if !ok {
		addError(result, codeAllowedToolsNotString, "Frontmatter 'allowed-tools' must be a string.", "SKILL.md", lineFor(lines, "allowed-tools"))
		return
	}
	if strings.TrimSpace(tools) == "" {
		addError(result, codeAllowedToolsEmpty, "Frontmatter 'allowed-tools' must not be empty.", "SKILL.md", lineFor(lines, "allowed-tools"))
		return
	}
	_ = strings.Fields(tools)
}

func mappingRoot(node *yaml.Node) (*yaml.Node, error) {
	if node.Kind == yaml.DocumentNode && len(node.Content) > 0 {
		node = node.Content[0]
	}
	if node.Kind != yaml.MappingNode {
		return nil, errors.New("not mapping")
	}
	return node, nil
}

func mapKeyLines(node *yaml.Node, offset int) map[string]int {
	lines := make(map[string]int)
	for i := 0; i < len(node.Content)-1; i += 2 {
		keyNode := node.Content[i]
		if keyNode.Kind == yaml.ScalarNode {
			lines[keyNode.Value] = keyNode.Line + offset - 1
		}
	}
	return lines
}

func lineFor(lines map[string]int, key string) int {
	if line, ok := lines[key]; ok {
		return line
	}
	return 0
}

func collectUnknownKeys(node *yaml.Node) []string {
	unknown := make([]string, 0)
	for i := 0; i < len(node.Content)-1; i += 2 {
		keyNode := node.Content[i]
		if keyNode.Kind != yaml.ScalarNode {
			continue
		}
		key := keyNode.Value
		if _, ok := knownKeys[key]; !ok {
			unknown = append(unknown, key)
		}
	}
	sort.Strings(unknown)
	return unknown
}

func checkOptionalDir(root, name, errCode, warnCode string, result *Result, opts Options) {
	path := filepath.Join(root, name)
	info, err := os.Stat(path)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return
		}
		return
	}
	if !info.IsDir() {
		addError(result, errCode, fmt.Sprintf("%s must be a directory if present.", name), name, 0)
		return
	}
	entries, err := os.ReadDir(path)
	if err != nil {
		return
	}
	if len(entries) == 0 {
		addWarning(result, opts, warnCode, fmt.Sprintf("%s directory is empty.", name), name, 0)
	}
}

func scanReferences(root, body string, result *Result, opts Options) {
	if opts.NoWarn {
		return
	}
	seen := make(map[string]struct{})

	for _, match := range linkPattern.FindAllStringSubmatch(body, -1) {
		if len(match) < 2 {
			continue
		}
		ref := strings.TrimSpace(match[1])
		collectRef(ref, seen)
	}

	lines := strings.Split(body, "\n")
	for _, line := range lines {
		matches := plainRefPattern.FindAllString(line, -1)
		for _, match := range matches {
			ref := strings.TrimSpace(match)
			ref = strings.TrimLeft(ref, " \t")
			ref = strings.TrimRight(ref, ".,;:)")
			collectRef(ref, seen)
		}
	}

	for ref := range seen {
		checkRef(root, ref, result, opts)
	}
}

func collectRef(ref string, seen map[string]struct{}) {
	if ref == "" {
		return
	}
	if strings.HasPrefix(ref, "http://") || strings.HasPrefix(ref, "https://") {
		return
	}
	if strings.HasPrefix(ref, "/") || strings.HasPrefix(ref, "#") {
		return
	}
	seen[ref] = struct{}{}
}

func checkRef(root, ref string, result *Result, opts Options) {
	if hasDotDot(ref) {
		addWarning(result, opts, codeRefContainsDotDot, fmt.Sprintf("Reference '%s' contains '..' path segments.", ref), "SKILL.md", 0)
	}
	trimmed := strings.TrimPrefix(ref, "./")
	if strings.Count(trimmed, "/") > 1 {
		addWarning(result, opts, codeRefTooDeep, fmt.Sprintf("Reference '%s' is nested deeper than one level.", ref), "SKILL.md", 0)
	}

	if !opts.CheckRefsExist {
		return
	}

	refPath := filepath.FromSlash(ref)
	target := filepath.Clean(filepath.Join(root, refPath))
	if !isWithinRoot(root, target) {
		addWarning(result, opts, codeRefEscapesRoot, fmt.Sprintf("Reference '%s' resolves outside the skill directory.", ref), "SKILL.md", 0)
		return
	}

	info, err := os.Lstat(target)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			addWarning(result, opts, codeRefMissingFile, fmt.Sprintf("Reference '%s' does not exist.", ref), "SKILL.md", 0)
		}
		return
	}

	if info.Mode()&os.ModeSymlink != 0 && !opts.FollowSymlinks {
		resolved, err := filepath.EvalSymlinks(target)
		if err != nil {
			addWarning(result, opts, codeRefMissingFile, fmt.Sprintf("Reference '%s' could not be resolved.", ref), "SKILL.md", 0)
			return
		}
		if !isWithinRoot(root, resolved) {
			addWarning(result, opts, codeRefEscapesRoot, fmt.Sprintf("Reference '%s' resolves outside the skill directory.", ref), "SKILL.md", 0)
			return
		}
	}
}

func hasDotDot(path string) bool {
	for _, part := range strings.Split(path, "/") {
		if part == ".." {
			return true
		}
	}
	return false
}

func isWithinRoot(root, target string) bool {
	rel, err := filepath.Rel(root, target)
	if err != nil {
		return false
	}
	if rel == "." {
		return true
	}
	sep := string(os.PathSeparator)
	if rel == ".." || strings.HasPrefix(rel, ".."+sep) {
		return false
	}
	return true
}

func addError(result *Result, code, message, file string, line int) {
	result.Errors = append(result.Errors, Finding{
		Level:   LevelError,
		Code:    code,
		Message: message,
		File:    file,
		Line:    line,
	})
}

func addWarning(result *Result, opts Options, code, message, file string, line int) {
	if opts.NoWarn {
		return
	}
	result.Warnings = append(result.Warnings, Finding{
		Level:   LevelWarning,
		Code:    code,
		Message: message,
		File:    file,
		Line:    line,
	})
}

func finalizeResult(result *Result, opts Options) {
	sortFindings(result.Errors)
	sortFindings(result.Warnings)

	valid := len(result.Errors) == 0
	if opts.Strict && len(result.Warnings) > 0 {
		valid = false
	}
	result.Valid = valid
}

func sortFindings(findings []Finding) {
	sort.Slice(findings, func(i, j int) bool {
		if findings[i].File != findings[j].File {
			return findings[i].File < findings[j].File
		}
		li := findings[i].Line
		lj := findings[j].Line
		if li == 0 {
			li = 1<<30
		}
		if lj == 0 {
			lj = 1<<30
		}
		if li != lj {
			return li < lj
		}
		return findings[i].Code < findings[j].Code
	})
}
