# sklint

> Validate Agent Skills against the official Agent Skills specification — fast, strict, and CI-ready.

`sklint` checks that a skill directory complies with the Agent Skills open standard.  
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

Prebuilt binaries for Linux, macOS, and Windows are available on the **Releases** page.

> macOS note: If you download a binary directly, you may need:
> ```bash
> xattr -dr com.apple.quarantine ./sklint
> ```

---

## ⚡ Quick Start

Validate a skill directory:

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

## 🧪 CI Usage

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

## ✅ What sklint Validates

### Required structure
- `SKILL.md` present
- Optional `scripts/`, `references/`, `assets/` directories validated if present

### Frontmatter rules
- Proper `---` delimiters
- Valid YAML
- Required fields: `name`, `description`

### Field constraints
- `name` format, length, and directory match
- `description` length
- `compatibility`, `license`
- `metadata` must be string → string
- `allowed-tools` format

### Best-practice warnings
- Empty Markdown body
- Very long `SKILL.md`
- Unknown top-level keys
- Empty optional directories
- Suspicious or missing relative file references

---

## 📦 Example JSON Output

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

## 🔧 CLI Options

```bash
sklint [options] <path>
```

Options:

- `--follow-symlinks`: Follow symlinks
- `--format text|json`: Output format: text or json (default "text")
- `--no-warn`: Suppress warnings
- `--strict`: Treat warnings as errors
- `--output <file>`: Write report to file

---

## 🎯 Why sklint?

- Designed specifically for the Agent Skills specification  
- Safe by default (no unsafe symlink traversal)  
- Works locally and in CI  
- Zero dependencies at runtime  
- Cross-platform  
