package main

import (
	"flag"
	"fmt"
	"os"

	"sklint/internal/report"
	"sklint/pkg/validator"
)

func main() {
	var (
		format        string
		strict        bool
		noWarn        bool
		output        string
		schemaVersion int
		followLinks   bool
	)

	flag.StringVar(&format, "format", "text", "Output format: text or json")
	flag.BoolVar(&strict, "strict", false, "Treat warnings as errors")
	flag.BoolVar(&noWarn, "no-warn", false, "Suppress warnings")
	flag.StringVar(&output, "output", "", "Write report to file")
	flag.IntVar(&schemaVersion, "schema-version", 1, "Schema version")
	flag.BoolVar(&followLinks, "follow-symlinks", false, "Follow symlinks")
	flag.Parse()

	if flag.NArg() != 1 {
		exitWithError("Usage: sklint <path>")
	}
	if schemaVersion != 1 {
		exitWithError(fmt.Sprintf("Unsupported schema version: %d", schemaVersion))
	}
	if format != "text" && format != "json" {
		exitWithError(fmt.Sprintf("Unsupported format: %s", format))
	}

	path := flag.Arg(0)
	opts := validator.Options{
		Strict:         strict,
		NoWarn:         noWarn,
		FollowSymlinks: followLinks,
		SchemaVersion:  schemaVersion,
		CheckRefsExist: true,
	}

	result, err := validator.ValidateSkill(path, opts)
	if err != nil {
		exitWithError(err.Error())
	}

	var outputBytes []byte
	if format == "json" {
		outputBytes, err = report.RenderJSON(result)
		if err != nil {
			exitWithError(err.Error())
		}
		outputBytes = append(outputBytes, '\n')
	} else {
		outputBytes = []byte(report.RenderText(result))
	}

	if output != "" {
		if err := os.WriteFile(output, outputBytes, 0o644); err != nil {
			exitWithError(err.Error())
		}
	} else {
		if _, err := os.Stdout.Write(outputBytes); err != nil {
			exitWithError(err.Error())
		}
	}

	if result.Valid {
		os.Exit(0)
	}
	os.Exit(1)
}

func exitWithError(message string) {
	_, _ = fmt.Fprintln(os.Stderr, message)
	os.Exit(2)
}
