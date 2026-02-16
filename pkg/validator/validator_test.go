package validator

import (
	"path/filepath"
	"runtime"
	"testing"
)

func fixturePath(t *testing.T, name string) string {
	t.Helper()
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("unable to locate test file path")
	}
	return filepath.Join(filepath.Dir(file), "..", "..", "testdata", name)
}

func TestValidMinimal(t *testing.T) {
	result, err := ValidateSkill(fixturePath(t, "valid-minimal"), Options{CheckRefsExist: true})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.Valid {
		t.Fatalf("expected valid result, got invalid: %#v", result)
	}
	if len(result.Errors) != 0 {
		t.Fatalf("expected no errors, got %d", len(result.Errors))
	}
}

func TestMissingSkillMD(t *testing.T) {
	result, err := ValidateSkill(fixturePath(t, "invalid-missing-skillmd"), Options{CheckRefsExist: true})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	assertFinding(t, result, LevelError, codeSkillMDMissing)
}

func TestFrontmatterErrors(t *testing.T) {
	cases := []struct {
		name string
		code string
	}{
		{"invalid-frontmatter-missing-start", codeFrontmatterStart},
		{"invalid-frontmatter-missing-end", codeFrontmatterEnd},
		{"invalid-frontmatter-invalid-yaml", codeFrontmatterInvalidYAML},
		{"invalid-frontmatter-not-mapping", codeFrontmatterNotMapping},
	}
	for _, tc := range cases {
		result, err := ValidateSkill(fixturePath(t, tc.name), Options{CheckRefsExist: true})
		if err != nil {
			t.Fatalf("unexpected error for %s: %v", tc.name, err)
		}
		assertFinding(t, result, LevelError, tc.code)
	}
}

func TestNameViolations(t *testing.T) {
	cases := []struct {
		name string
		code string
	}{
		{"invalid-name-uppercase", codeNameInvalidChars},
		{"invalid-name-start-hyphen", codeNameStartsWithHyphen},
		{"invalid-name-end-hyphen", codeNameEndsWithHyphen},
		{"invalid-name-consecutive-hyphen", codeNameConsecutiveHyphens},
		{"invalid-name-too-long", codeNameTooLong},
		{"invalid-name-too-short", codeNameTooShort},
		{"invalid-name-mismatch", codeNameMismatchDirectory},
	}
	for _, tc := range cases {
		result, err := ValidateSkill(fixturePath(t, tc.name), Options{CheckRefsExist: true})
		if err != nil {
			t.Fatalf("unexpected error for %s: %v", tc.name, err)
		}
		assertFinding(t, result, LevelError, tc.code)
	}
}

func TestDescriptionLength(t *testing.T) {
	cases := []struct {
		name string
		code string
	}{
		{"invalid-description-empty", codeDescriptionTooShort},
		{"invalid-description-too-long", codeDescriptionTooLong},
	}
	for _, tc := range cases {
		result, err := ValidateSkill(fixturePath(t, tc.name), Options{CheckRefsExist: true})
		if err != nil {
			t.Fatalf("unexpected error for %s: %v", tc.name, err)
		}
		assertFinding(t, result, LevelError, tc.code)
	}
}

func TestCompatibilityLength(t *testing.T) {
	cases := []struct {
		name string
		code string
	}{
		{"invalid-compatibility-empty", codeCompatibilityTooShort},
		{"invalid-compatibility-too-long", codeCompatibilityTooLong},
	}
	for _, tc := range cases {
		result, err := ValidateSkill(fixturePath(t, tc.name), Options{CheckRefsExist: true})
		if err != nil {
			t.Fatalf("unexpected error for %s: %v", tc.name, err)
		}
		assertFinding(t, result, LevelError, tc.code)
	}
}

func TestMetadataValidation(t *testing.T) {
	cases := []struct {
		name string
		code string
	}{
		{"invalid-metadata-not-object", codeMetadataNotObject},
		{"invalid-metadata-value-not-string", codeMetadataValueNotString},
	}
	for _, tc := range cases {
		result, err := ValidateSkill(fixturePath(t, tc.name), Options{CheckRefsExist: true})
		if err != nil {
			t.Fatalf("unexpected error for %s: %v", tc.name, err)
		}
		assertFinding(t, result, LevelError, tc.code)
	}
}

func TestAllowedToolsValidation(t *testing.T) {
	result, err := ValidateSkill(fixturePath(t, "invalid-allowed-tools-empty"), Options{CheckRefsExist: true})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	assertFinding(t, result, LevelError, codeAllowedToolsEmpty)
}

func TestReferenceWarnings(t *testing.T) {
	result, err := ValidateSkill(fixturePath(t, "reference-warnings"), Options{CheckRefsExist: true})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	assertFinding(t, result, LevelWarning, codeRefMissingFile)
	assertFinding(t, result, LevelWarning, codeRefTooDeep)
	assertFinding(t, result, LevelWarning, codeRefContainsDotDot)
	assertFinding(t, result, LevelWarning, codeRefEscapesRoot)
}

func TestLineCountWarning(t *testing.T) {
	result, err := ValidateSkill(fixturePath(t, "line-count-warning"), Options{CheckRefsExist: true})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	assertFinding(t, result, LevelWarning, codeSkillMDTooLongLines)
}

func TestStrictMode(t *testing.T) {
	result, err := ValidateSkill(fixturePath(t, "strict-warning-only"), Options{CheckRefsExist: true, Strict: true})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Valid {
		t.Fatalf("expected invalid in strict mode when warnings present")
	}
	if len(result.Warnings) == 0 {
		t.Fatalf("expected warnings")
	}
}

func assertFinding(t *testing.T, result Result, level FindingLevel, code string) {
	t.Helper()
	for _, finding := range result.Errors {
		if finding.Level == level && finding.Code == code {
			return
		}
	}
	for _, finding := range result.Warnings {
		if finding.Level == level && finding.Code == code {
			return
		}
	}
	t.Fatalf("expected finding %s/%s not found", level, code)
}
