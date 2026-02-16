package report

import (
	"strings"
	"testing"

	"github.com/sven1103-agent/sklint/pkg/validator"
)

func TestRenderText(t *testing.T) {
	result := validator.Result{
		Path:  "/tmp/skill",
		Valid: false,
		Errors: []validator.Finding{
			{Level: validator.LevelError, Code: "ERR", Message: "bad", File: "SKILL.md", Line: 3},
		},
		Warnings: []validator.Finding{
			{Level: validator.LevelWarning, Code: "WARN", Message: "note"},
		},
	}
	out := RenderText(result)
	if !strings.Contains(out, "Errors") || !strings.Contains(out, "Warnings") {
		t.Fatalf("expected sections, got %q", out)
	}
	if !strings.Contains(out, "ERR") || !strings.Contains(out, "WARN") {
		t.Fatalf("expected findings, got %q", out)
	}
	if !strings.Contains(out, "1 errors, 1 warnings - INVALID") {
		t.Fatalf("unexpected summary: %q", out)
	}
}
