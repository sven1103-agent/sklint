# sklint

`sklint` validates an Agent Skills "skill directory" against the Agent Skills open specification. It checks `SKILL.md` frontmatter rules, required fields, optional directories, and best-practice warnings.

**What it validates**
- Skill directory structure (`SKILL.md`, optional `scripts/`, `references/`, `assets/`)
- Frontmatter delimiters, YAML validity, and required fields
- `name`, `description`, `compatibility`, `license`, `metadata`, `allowed-tools` constraints
- Best-practice warnings (empty body, long file, unknown keys, empty optional directories)
- Markdown file reference hygiene

**Installation**
- Go install:
  - `go install github.com/your-org/agentskills-validator/cmd/sklint@latest`
- Or download a binary from Releases (if CI is enabled).

**Usage**
- `sklint ./my-skill`
- `sklint --format json --output report.json ./my-skill`
- `sklint --strict ./my-skill`

**Output (text)**
```
Errors
- NAME_MISMATCH_DIRECTORY SKILL.md:3 Frontmatter name 'pdf-processing' must match directory name 'pdf_processing'.

Warnings
- SKILL_MD_TOO_LONG_LINES SKILL.md SKILL.md is 742 lines; recommended under 500 lines.

1 errors, 1 warnings - INVALID
```

**Output (json)**
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
  "warnings": [
    {
      "level": "warning",
      "code": "SKILL_MD_TOO_LONG_LINES",
      "message": "SKILL.md is 742 lines; recommended under 500 lines.",
      "file": "SKILL.md"
    }
  ]
}
```

**Exit codes**
- `0` valid (no errors; warnings allowed unless `--strict`)
- `1` validation errors present (or warnings in `--strict`)
- `2` runtime or usage error
