package report

import (
	"fmt"
	"strings"

	"sklint/pkg/validator"
)

func RenderText(result validator.Result) string {
	var b strings.Builder
	if len(result.Errors) > 0 {
		b.WriteString("Errors\n")
		for _, finding := range result.Errors {
			b.WriteString(formatFinding(finding))
		}
	}
	if len(result.Warnings) > 0 {
		if b.Len() > 0 {
			b.WriteString("\n")
		}
		b.WriteString("Warnings\n")
		for _, finding := range result.Warnings {
			b.WriteString(formatFinding(finding))
		}
	}
	if b.Len() > 0 {
		b.WriteString("\n")
	}

	status := "VALID"
	if !result.Valid {
		status = "INVALID"
	}
	b.WriteString(fmt.Sprintf("%d errors, %d warnings - %s\n", len(result.Errors), len(result.Warnings), status))
	return b.String()
}

func formatFinding(f validator.Finding) string {
	location := ""
	if f.File != "" {
		location = f.File
		if f.Line > 0 {
			location = fmt.Sprintf("%s:%d", f.File, f.Line)
		}
		location = " " + location
	}
	return fmt.Sprintf("- %s%s %s\n", f.Code, location, f.Message)
}
