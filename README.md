# sklint

[![CI](https://github.com/sven1103-agent/sklint/actions/workflows/release.yml/badge.svg)](https://github.com/sven1103-agent/sklint/actions)
[![Go Version](https://img.shields.io/badge/go-1.21+-blue.svg)](https://go.dev/)
[![License](https://img.shields.io/badge/license-AGPL3-green.svg)](LICENSE)

> Validate Agent Skills against the official Agent Skills specification — fast, strict, and CI-ready.

`sklint` checks that a skill directory complies with the [Agent Skills open specification](https://agentskills.io/specification).  
It validates structure, frontmatter rules, required fields, and common best-practice issues.

---

## Install

### From source (this repo)

```bash
go install ./cmd/sklint
```

Install a specific tagged release (tags use `vYYYY-MM-DD`, e.g. `v2026-02-16`):

```bash
git checkout v2026-02-16
go install ./cmd/sklint
```

### Install via `go install`

```bash
go install github.com/sven1103-agent/sklint/cmd/sklint@latest
```

Install a specific release:

```bash
go install github.com/sven1103-agent/sklint/cmd/sklint@v2026-02-16
```

### Binary download (GitHub Releases)

Prebuilt binaries for Linux, macOS, and Windows are available on the [Releases page](https://github.com/sven1103-agent/sklint/releases).

> macOS note: If you download a binary directly, you may need:
> ```bash
> xattr -dr com.apple.quarantine ./sklint
> ```

---

## Quick Start

Validate a skill directory (the folder containing `SKILL.md`):

```bash
sklint ./my-skill
```

Example:

```bash
sklint ./skills/pdf-processing
```

Example output:

```
Errors
- NAME_MISMATCH_DIRECTORY SKILL.md:3 Frontmatter name 'pdf-processing' must match directory name 'pdf_processing'.

1 errors, 0 warnings - INVALID
```

---

## Valid Skill Example

A minimal valid skill directory:

```
my-skill/
├── SKILL.md
├── scripts/       (optional)
├── references/    (optional)
└── assets/        (optional)
```

Example `SKILL.md`:

```yaml
---
name: my-skill
description: A skill that does something useful.
license: MIT
compatibility: Works with Claude, GPT-4, and similar LLMs.
metadata:
  author: Jane Doe
  version: "1.0"
allowed-tools: bash read write
---

# My Skill

This skill helps with...
```

Successful validation output:

```
0 errors, 0 warnings - VALID
```

---

## CI Usage

Strict mode fails on warnings:

```bash
sklint --strict .
```

JSON output for pipelines:

```bash
sklint --format json --output report.json .
```

Exit codes:

| Code | Meaning |
|------|---------|
| 0 | Valid (no errors; warnings allowed unless `--strict`) |
| 1 | Validation errors (or warnings in strict mode) |
| 2 | Runtime / usage error |

---

## What sklint Validates

### Required structure
- `SKILL.md` present
- Optional `scripts/`, `references/`, `assets/` directories:
  - Must be directories (not files) if present
  - Warns if directory exists but is empty

### Frontmatter rules
- Proper `---` delimiters
- Valid YAML
- Required fields: `name`, `description`

### Field constraints

| Field | Required | Type | Constraints |
|-------|----------|------|-------------|
| `name` | Yes | string | 1-64 characters; lowercase `a-z`, digits `0-9`, and hyphens `-` only; cannot start or end with hyphen; no consecutive hyphens `--`; must match the directory name |
| `description` | Yes | string | 1-1024 characters |
| `license` | No | string | Any string (e.g., `MIT`, `Apache-2.0`) |
| `compatibility` | No | string | 1-500 characters |
| `metadata` | No | object | Keys and values must both be strings |
| `allowed-tools` | No | string | Whitespace-delimited tool names; cannot be empty if present |

### Best-practice warnings
- Empty Markdown body
- Very long `SKILL.md`
- Unknown top-level keys
- Empty optional directories
- Suspicious or missing relative file references

---

## Example JSON Output

```json
{
  "path": "/abs/path/to/skill",
  "valid": false,
  "errors": [
    {
      "level": "error",
      "code": "NAME_MISMATCH_DIRECTORY",
      "message": "Frontmatter name 'pdf-processing' must match directory name 'pdf_processing'.",
      "file": "SKILL.md",
      "line": 3
    }
  ],
  "warnings": []
}
```

---

## CLI Options

```bash
sklint [options] <skill-directory>
```

Where `<skill-directory>` is the path to the folder containing `SKILL.md`.

Run `sklint --help` to see usage information.

Options:

- `--follow-symlinks`: Follow symlinks
- `--format text|json`: Output format: text or json (default "text")
- `--no-warn`: Suppress warnings
- `--strict`: Treat warnings as errors
- `--output <file>`: Write report to file

---

## Error and Warning Codes

### Structure Errors

| Code | Description |
|------|-------------|
| `PATH_NOT_FOUND` | The specified path does not exist |
| `PATH_NOT_DIRECTORY` | The specified path is not a directory |
| `SKILL_MD_MISSING` | No `SKILL.md` file found in the directory |
| `SKILL_MD_NOT_FILE` | `SKILL.md` exists but is a directory |
| `SKILL_MD_SYMLINK_ESCAPES_ROOT` | `SKILL.md` symlink points outside the skill directory |

### Frontmatter Errors

| Code | Description |
|------|-------------|
| `FRONTMATTER_START_MISSING` | File does not begin with `---` |
| `FRONTMATTER_END_MISSING` | No closing `---` delimiter found |
| `FRONTMATTER_EMPTY` | No content between `---` delimiters |
| `FRONTMATTER_INVALID_YAML` | YAML syntax error |
| `FRONTMATTER_NOT_MAPPING` | YAML is not a key-value mapping |

### Name Field Errors

| Code | Description |
|------|-------------|
| `NAME_MISSING` | Required `name` field not present |
| `NAME_NOT_STRING` | `name` value is not a string |
| `NAME_TOO_SHORT` | `name` is empty (0 characters) |
| `NAME_TOO_LONG` | `name` exceeds 64 characters |
| `NAME_INVALID_CHARS` | `name` contains invalid characters |
| `NAME_STARTS_WITH_HYPHEN` | `name` begins with `-` |
| `NAME_ENDS_WITH_HYPHEN` | `name` ends with `-` |
| `NAME_CONSECUTIVE_HYPHENS` | `name` contains `--` |
| `NAME_MISMATCH_DIRECTORY` | `name` does not match the directory name |

### Description Field Errors

| Code | Description |
|------|-------------|
| `DESCRIPTION_MISSING` | Required `description` field not present |
| `DESCRIPTION_NOT_STRING` | `description` value is not a string |
| `DESCRIPTION_TOO_SHORT` | `description` is empty |
| `DESCRIPTION_TOO_LONG` | `description` exceeds 1024 characters |

### Other Field Errors

| Code | Description |
|------|-------------|
| `COMPATIBILITY_NOT_STRING` | `compatibility` is not a string |
| `COMPATIBILITY_TOO_SHORT` | `compatibility` is empty |
| `COMPATIBILITY_TOO_LONG` | `compatibility` exceeds 500 characters |
| `LICENSE_NOT_STRING` | `license` is not a string |
| `METADATA_NOT_OBJECT` | `metadata` is not a key-value object |
| `METADATA_VALUE_NOT_STRING` | `metadata` contains non-string values |
| `ALLOWED_TOOLS_NOT_STRING` | `allowed-tools` is not a string |
| `ALLOWED_TOOLS_EMPTY` | `allowed-tools` is empty or whitespace-only |

### Warnings

| Code | Description |
|------|-------------|
| `SKILL_MD_SYMLINK` | `SKILL.md` is a symlink (informational) |
| `SKILL_MD_TOO_LONG_LINES` | `SKILL.md` exceeds 500 lines |
| `SKILL_MD_MISSING_BODY` | No content after frontmatter |
| `UNKNOWN_TOP_LEVEL_KEY` | Unrecognized keys in frontmatter |
| `SCRIPTS_DIR_EMPTY` | `scripts/` directory exists but is empty |
| `REFERENCES_DIR_EMPTY` | `references/` directory exists but is empty |
| `ASSETS_DIR_EMPTY` | `assets/` directory exists but is empty |
| `REF_CONTAINS_DOTDOT` | Reference path contains `..` |
| `REF_TOO_DEEP` | Reference path is more than one level deep |
| `REF_MISSING_FILE` | Referenced file does not exist |
| `REF_ESCAPES_ROOT` | Reference resolves outside skill directory |

---

## Using as a Go Library

Import the validator package to embed validation in your own tools:

```go
package main

import (
    "fmt"
    "log"

    "github.com/sven1103-agent/sklint/pkg/validator"
)

func main() {
    result, err := validator.ValidateSkill("./my-skill", validator.Options{
        Strict:         false,  // treat warnings as errors
        NoWarn:         false,  // suppress warnings
        FollowSymlinks: false,  // follow symlinks outside root
        CheckRefsExist: true,   // verify referenced files exist
    })
    if err != nil {
        log.Fatal(err)  // runtime error (I/O, permissions)
    }

    if result.Valid {
        fmt.Println("Skill is valid!")
    } else {
        for _, e := range result.Errors {
            fmt.Printf("[%s] %s:%d - %s\n", e.Code, e.File, e.Line, e.Message)
        }
    }
}
```

---

## Troubleshooting

### macOS: "cannot be opened because the developer cannot be verified"

If you downloaded a binary directly:
```bash
xattr -dr com.apple.quarantine ./sklint
```

### NAME_MISMATCH_DIRECTORY error

The `name` field in SKILL.md must exactly match the directory name:
```
my-skill/           <- directory name
└── SKILL.md
    ---
    name: my-skill  <- must match
    ---
```

### Permission denied

Ensure you have read access to all files in the skill directory:
```bash
chmod -R u+r ./my-skill
```

---

## Why sklint?

- **Specification-compliant**: Validates against the official Agent Skills spec
- **Actionable output**: Error codes with line numbers and clear fix instructions
- **CI-ready**: JSON output, strict mode, and meaningful exit codes
- **Secure defaults**: Won't follow symlinks outside the skill directory
- **Embeddable**: Use as a CLI or import as a Go library  

---

## License

This project is licensed under the [AGPL-3.0 License](LICENSE).

---

## Contributing

Contributions are welcome! Please:

1. Open an issue to discuss proposed changes
2. Fork the repository and create a feature branch
3. Run tests with `go test ./...`
4. Submit a pull request

See the test fixtures in `testdata/` for examples of valid and invalid skills.
