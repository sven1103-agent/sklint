package parse

import (
	"strings"
	"testing"
)

func TestParseFrontmatterValid(t *testing.T) {
	input := "---\nname: test\n---\nbody\n"
	fm, err := ParseFrontmatter(strings.NewReader(input))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if fm.YAML != "name: test" {
		t.Fatalf("unexpected yaml: %q", fm.YAML)
	}
	if fm.Body != "body" {
		t.Fatalf("unexpected body: %q", fm.Body)
	}
	if fm.LineCount != 4 {
		t.Fatalf("unexpected line count: %d", fm.LineCount)
	}
	if fm.YAMLStartLine != 2 {
		t.Fatalf("unexpected yaml start line: %d", fm.YAMLStartLine)
	}
}

func TestParseFrontmatterBOM(t *testing.T) {
	input := "\xEF\xBB\xBF---\nname: test\n---\nbody\n"
	fm, err := ParseFrontmatter(strings.NewReader(input))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if fm.YAML != "name: test" {
		t.Fatalf("unexpected yaml: %q", fm.YAML)
	}
}

func TestParseFrontmatterErrors(t *testing.T) {
	cases := []struct {
		name  string
		input string
		want  error
	}{
		{"missing-start", "name: test\n---\nbody\n", ErrFrontmatterStartMissing},
		{"missing-end", "---\nname: test\nbody\n", ErrFrontmatterEndMissing},
		{"empty-yaml", "---\n---\nbody\n", ErrFrontmatterEmpty},
	}
	for _, tc := range cases {
		_, err := ParseFrontmatter(strings.NewReader(tc.input))
		if err != tc.want {
			t.Fatalf("%s: expected %v, got %v", tc.name, tc.want, err)
		}
	}
}
