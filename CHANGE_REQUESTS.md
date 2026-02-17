# Documentation Change Requests for sklint

This document outlines suggested improvements to the README.md documentation,
prioritized by importance.

## Priority Levels

- **P0 (Critical)**: Missing information that prevents users from succeeding
- **P1 (High)**: Important gaps that significantly impact user experience
- **P2 (Medium)**: Improvements that enhance clarity and completeness
- **P3 (Low)**: Nice-to-have polish and consistency fixes

---

## P0 - Critical

### CR-001: Add Link to Agent Skills Specification

**Location**: Introduction (lines 3-6)

**Issue**: The README mentions "Agent Skills open standard" but provides no link
to the specification. Users cannot verify what rules are being enforced or
understand the broader context.

**Suggested Change**:

Replace:
```markdown
`sklint` checks that a skill directory complies with the Agent Skills open standard.
```

With:
```markdown
`sklint` checks that a skill directory complies with the [Agent Skills open specification](https://agentskills.io/specification).
```

---

### CR-002: Add Valid SKILL.md Example

**Location**: New section after "Quick Start"

**Issue**: Users have no reference for what a valid skill directory looks like.
The only example shown demonstrates an error case, which doesn't help users
understand the expected format.

**Suggested Change**: Add a new section:

```markdown
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
```

---

## P1 - High Priority

### CR-003: Document Complete Field Constraints

**Location**: "What sklint Validates" section (lines 108-114)

**Issue**: Field constraints are vague. Users cannot know the actual rules
without reading the source code. For example, "name format, length" doesn't
specify:
- Allowed characters
- Length limits
- Naming rules

**Suggested Change**: Replace the brief list with a detailed table:

```markdown
### Field constraints

| Field | Required | Type | Constraints |
|-------|----------|------|-------------|
| `name` | Yes | string | 1-64 characters; lowercase `a-z`, digits `0-9`, and hyphens `-` only; cannot start or end with hyphen; no consecutive hyphens `--`; must match the directory name |
| `description` | Yes | string | 1-1024 characters |
| `license` | No | string | Any string (e.g., `MIT`, `Apache-2.0`) |
| `compatibility` | No | string | 1-500 characters |
| `metadata` | No | object | Keys and values must both be strings |
| `allowed-tools` | No | string | Whitespace-delimited tool names; cannot be empty if present |
```

---

### CR-004: Add Error Code Reference

**Location**: New section or appendix

**Issue**: Users see codes like `NAME_MISMATCH_DIRECTORY` in output but have no
documentation explaining what each code means or how to fix it.

**Suggested Change**: Add a new section:

```markdown
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
```

---

### CR-005: Add Library Usage Example

**Location**: New section "Using as a Go Library"

**Issue**: The project provides a public `pkg/validator` package but the README
shows no usage example. Developers wanting to embed the validator have no
starting point.

**Suggested Change**: Add a new section:

```markdown
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
```

---

### CR-006: Add Link to GitHub Releases

**Location**: "Binary download" section (line 39)

**Issue**: The text says binaries are "available on the Releases page" but
doesn't provide a clickable link.

**Suggested Change**:

Replace:
```markdown
Prebuilt binaries for Linux, macOS, and Windows are available on the **Releases** page.
```

With:
```markdown
Prebuilt binaries for Linux, macOS, and Windows are available on the [Releases page](https://github.com/sven1103-agent/sklint/releases).
```

---

## P2 - Medium Priority

### CR-007: Clarify What "path" Means

**Location**: Quick Start and CLI Options sections

**Issue**: The `<path>` argument could be confused with the path to SKILL.md
file rather than the skill directory.

**Suggested Change**:

In Quick Start, change:
```markdown
Validate a skill directory:
```

To:
```markdown
Validate a skill directory (the folder containing SKILL.md):
```

In CLI Options, change:
```markdown
sklint [options] <path>
```

To:
```markdown
sklint [options] <skill-directory>
```

And add below:
```markdown
Where `<skill-directory>` is the path to the folder containing `SKILL.md`.
```

---

### CR-008: Explain Optional Directory Validation

**Location**: "What sklint Validates" section (line 101)

**Issue**: "Optional directories validated if present" doesn't explain what
validation actually occurs on these directories.

**Suggested Change**:

Replace:
```markdown
- Optional `scripts/`, `references/`, `assets/` directories validated if present
```

With:
```markdown
- Optional `scripts/`, `references/`, `assets/` directories:
  - Must be directories (not files) if present
  - Warns if directory exists but is empty
```

---

### CR-009: Document --help Flag Behavior

**Location**: CLI Options section

**Issue**: Users expect `--help` or `-h` flags. Current behavior when running
without arguments is not documented.

**Suggested Change**: Add to CLI Options:

```markdown
Run `sklint` without arguments or with `--help` to see usage information.
```

Note: If `--help` is not implemented, this is also a feature request.

---

### CR-010: Add License Section

**Location**: End of README

**Issue**: A LICENSE file exists in the repository but isn't referenced in the
README. Users should know the license terms at a glance.

**Suggested Change**: Add before or after "Why sklint?":

```markdown
## License

This project is licensed under the [MIT License](LICENSE).
```

---

### CR-011: Add Contributing Section

**Location**: End of README

**Issue**: Open source project has no guidance for contributors.

**Suggested Change**: Add a new section:

```markdown
## Contributing

Contributions are welcome! Please:

1. Open an issue to discuss proposed changes
2. Fork the repository and create a feature branch
3. Run tests with `go test ./...`
4. Submit a pull request

See the test fixtures in `testdata/` for examples of valid and invalid skills.
```

---

## P3 - Low Priority (Polish)

### CR-012: Add README Badges

**Location**: Top of README, after title

**Issue**: No visual indicators for build status, Go version compatibility, or
license type. Badges help users quickly assess project health.

**Suggested Change**: Add after the title:

```markdown
# sklint

[![CI](https://github.com/sven1103-agent/sklint/actions/workflows/release.yml/badge.svg)](https://github.com/sven1103-agent/sklint/actions)
[![Go Version](https://img.shields.io/badge/go-1.21+-blue.svg)](https://go.dev/)
[![License](https://img.shields.io/badge/license-MIT-green.svg)](LICENSE)
```

---

### CR-013: Fix Emoji Inconsistency in Section Headers

**Location**: All section headers

**Issue**: Inconsistent emoji usage across sections:
- **With emoji**: Quick Start, CI Usage, What sklint Validates, Example JSON
  Output, CLI Options, Why sklint?
- **Without emoji**: Install

**Suggested Change**: Either:

Option A - Add emoji to Install:
```markdown
## 📥 Install
```

Option B - Remove all emojis for a cleaner, more professional look.

Recommendation: Option B (remove emojis) for consistency with typical CLI tool
documentation.

---

### CR-014: Strengthen "Why sklint?" Section

**Location**: Lines 161-167

**Issue**: The benefits listed are generic and apply to most Go CLI tools:
- "Cross-platform" - true of any Go binary
- "Zero dependencies at runtime" - true of any Go binary
- "Works locally and in CI" - true of any CLI tool

**Suggested Change**: Replace with more specific, compelling points:

```markdown
## Why sklint?

- **Specification-compliant**: Validates against the official Agent Skills spec
- **Actionable output**: Error codes with line numbers and clear fix instructions
- **CI-ready**: JSON output, strict mode, and meaningful exit codes
- **Secure defaults**: Won't follow symlinks outside the skill directory
- **Embeddable**: Use as a CLI or import as a Go library
```

---

### CR-015: Add Troubleshooting Section

**Location**: New section before "Why sklint?"

**Issue**: No guidance for common issues. The macOS quarantine note is buried
in the install section.

**Suggested Change**: Add a new section:

```markdown
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

## Summary

| Priority | Count | Description |
|----------|-------|-------------|
| P0 | 2 | Critical missing information |
| P1 | 4 | High-impact improvements |
| P2 | 5 | Clarity and completeness |
| P3 | 4 | Polish and consistency |

**Total**: 15 change requests

---

## Implementation Order

Recommended order for implementing these changes:

- [x] **CR-001** + **CR-006**: Add specification and releases links (quick wins)
- [x] **CR-002**: Add valid skill example (high user value)
- [x] **CR-003**: Document field constraints (essential reference)
- [x] **CR-004**: Add error code reference (essential reference)
- [x] **CR-005**: Add library usage example
- [ ] **CR-007** + **CR-008**: Clarify path and optional directories
- [ ] **CR-010** + **CR-011**: Add license and contributing sections
- [ ] **CR-009**: Document --help behavior
- [ ] **CR-012** - **CR-015**: Polish items (can be batched)
