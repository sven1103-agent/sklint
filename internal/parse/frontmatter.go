package parse

import (
	"bufio"
	"bytes"
	"errors"
	"io"
	"strings"
)

type Frontmatter struct {
	YAML          string
	Body          string
	LineCount     int
	YAMLStartLine int
}

var (
	ErrFrontmatterStartMissing = errors.New("frontmatter start missing")
	ErrFrontmatterEndMissing   = errors.New("frontmatter end missing")
	ErrFrontmatterEmpty        = errors.New("frontmatter empty")
)

func ParseFrontmatter(r io.Reader) (Frontmatter, error) {
	content, err := io.ReadAll(r)
	if err != nil {
		return Frontmatter{}, err
	}

	if bytes.HasPrefix(content, []byte{0xEF, 0xBB, 0xBF}) {
		content = content[3:]
	}

	scanner := bufio.NewScanner(bytes.NewReader(content))
	lines := make([]string, 0)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	if err := scanner.Err(); err != nil {
		return Frontmatter{}, err
	}

	lineCount := len(lines)
	if lineCount == 0 {
		return Frontmatter{}, ErrFrontmatterStartMissing
	}
	if lines[0] != "---" {
		return Frontmatter{}, ErrFrontmatterStartMissing
	}

	end := -1
	for i := 1; i < len(lines); i++ {
		if lines[i] == "---" {
			end = i
			break
		}
	}
	if end == -1 {
		return Frontmatter{}, ErrFrontmatterEndMissing
	}

	yamlLines := lines[1:end]
	yamlText := strings.Join(yamlLines, "\n")
	if strings.TrimSpace(yamlText) == "" {
		return Frontmatter{}, ErrFrontmatterEmpty
	}

	bodyLines := []string{}
	if end+1 < len(lines) {
		bodyLines = lines[end+1:]
	}
	bodyText := strings.Join(bodyLines, "\n")

	return Frontmatter{
		YAML:          yamlText,
		Body:          bodyText,
		LineCount:     lineCount,
		YAMLStartLine: 2,
	}, nil
}
