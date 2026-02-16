# Instruction prompt: Build an Agent Skills Specification Validator (Go CLI)

You are a coding agent. Build a **Go** command-line application that validates an **Agent Skills “skill directory”** against the **Agent Skills open specification** (agentskills.io/specification).

This prompt is written to be directly actionable: implement the CLI, the validator library, tests, fixtures, and (optionally) CI for cross-platform binaries.

---

## 0) Goals and non-goals

### Goals
- Validate a **single skill directory** (folder) that contains a required `SKILL.md` file and optional directories such as `scripts/`, `references/`, `assets/` (validate if present).
- Produce:
  - **Human-readable** output (`text`), and
  - **Machine-readable** output (`json`).
- Return a **non-zero exit code** if any **errors** are found.
- Support **warnings** (non-fatal by default; optionally fatal in strict mode).
- Provide a reusable Go library package for other tooling to embed.

### Non-goals
- Do not enforce the Markdown body structure beyond frontmatter rules (the spec says body has no format restrictions).
- Do not enforce semantic quality of description beyond length/type.

---

## 1) Repository layout (Go)

If the current working directory is already a Git repository, do not add a root folder for the project. Take the current one
as root directory for the project. 

The agent MUST not override existing files.

Create a standalone repo named e.g. `agentskills-validator` with this structure:

```
/cmd/sklint/           # CLI entrypoint
  main.go
/pkg/validator/                # public validation library
  validator.go
  rules.go
  types.go
/internal/parse/               # internal helpers (YAML frontmatter parsing)
  frontmatter.go
/internal/report/              # text & JSON renderers
  json.go
  text.go
/testdata/                     # fixture skill directories for tests
  valid-minimal/
  invalid-missing-skillmd/
  invalid-name-uppercase/
  ...
go.mod
README.md
```

Conventions:
- Keep the **validator engine** in `pkg/validator`.
- Keep filesystem and parsing helpers in `internal/...`.
- CLI should be thin: parse flags → call validator → print report → exit code.

---

## 2) CLI UX requirements

Binary name: `sklint`

### Command
```
sklint <path>
```

### Flags
- `--format text|json` (default `text`)
- `--strict` (treat warnings as errors)
- `--no-warn` (suppress warnings)
- `--output <file>` (write report to a file instead of stdout)
- `--schema-version <v>` (default `1`; reserved for future compatibility)
- Optional:
  - `--follow-symlinks` (default false; security-safe default)

### Exit codes
- `0` = valid (no errors; warnings allowed unless `--strict`)
- `1` = validation errors present (or warnings in `--strict`)
- `2` = tool/runtime failure (I/O, permissions, invalid args)

### Examples
```
sklint ./my-skill
sklint --format json --output report.json ./my-skill
sklint --strict ./my-skill
```

---

## 3) Public library API (pkg/validator)

Implement:

```go
package validator

type Options struct {
    Strict         bool
    NoWarn         bool
    FollowSymlinks bool
    SchemaVersion  int
    CheckRefsExist bool // default true
}

type FindingLevel string
const (
    LevelError   FindingLevel = "error"
    LevelWarning FindingLevel = "warning"
)

type Finding struct {
    Level   FindingLevel `json:"level"`
    Code    string       `json:"code"`
    Message string       `json:"message"`
    File    string       `json:"file,omitempty"`
    Line    int          `json:"line,omitempty"`
}

type Result struct {
    Path     string    `json:"path"`
    Valid    bool      `json:"valid"`
    Errors   []Finding `json:"errors,omitempty"`
    Warnings []Finding `json:"warnings,omitempty"`
}

func ValidateSkill(path string, opts Options) (Result, error)
```

Notes:
- `Result.Valid` is true only when there are **no errors** (and no warnings if strict).
- `ValidateSkill` returns a Go `error` only for runtime failures, not for validation failures.
- `Finding.Line` should be best-effort (0 if unknown). Use YAML parser node line info when feasible.

---

## 4) What is a “skill directory”

A skill directory is a directory whose root contains:
- `SKILL.md` (required)

Optional subdirectories (validate if present):
- `scripts/`
- `references/`
- `assets/`

### Skill directory checks (ERRORS)
- The provided path exists and is a directory.
- `<root>/SKILL.md` exists and is a file.
- If optional directories exist, they must be directories (not regular files).
- If `SKILL.md` is a symlink:
  - If `--follow-symlinks=false`, read it only if it resolves inside the skill directory, otherwise error.
  - Always add a **warning** that `SKILL.md` is a symlink.

Security:
- Default behavior must **not follow symlinks** that escape the skill root when validating referenced files.

---

## 5) Parse SKILL.md: YAML frontmatter + Markdown body

`SKILL.md` must contain YAML frontmatter followed by Markdown.

### Frontmatter delimiter rules (ERRORS)
- File must begin with a line exactly `---`.
- YAML ends at the next line exactly `---`.
- Must contain at least one YAML key/value pair between delimiters (non-empty mapping).
- YAML must parse successfully.
- YAML top-level must be a mapping/object.

After closing delimiter, body may be empty (but warn; see §8).

### Implementation approach
- Read file as UTF-8 (accept BOM if present).
- Parse frontmatter manually:
  - Locate first line (`---`).
  - Collect YAML until second `---`.
  - Collect remainder as Markdown body (string).

For YAML parsing use:
- `gopkg.in/yaml.v3`

Bonus: Use the AST (`yaml.Node`) to capture line numbers for known keys.

---

## 6) Frontmatter fields and constraints

### Required fields
- `name` (required)
- `description` (required)

### Optional fields
- `license` (optional string)
- `compatibility` (optional string length 1–500)
- `metadata` (optional object: string→string only)
- `allowed-tools` (optional string; validate tokens)

Unknown top-level keys:
- Do not error.
- Emit a warning listing unknown keys (unless `--no-warn`).

---

## 7) Validation rules (ERRORS)

### 7.1 `name` validation (ERRORS)
Rules:
- Type: string.
- Length: **1–64** characters.
- Allowed characters: lowercase letters `a-z`, digits `0-9`, hyphen `-` only.
- Must not start or end with `-`.
- Must not contain consecutive hyphens `--`.
- Must match the **parent directory name** exactly (basename of skill path).

Suggested error codes:
- `NAME_MISSING`
- `NAME_NOT_STRING`
- `NAME_TOO_SHORT`
- `NAME_TOO_LONG`
- `NAME_INVALID_CHARS`
- `NAME_STARTS_WITH_HYPHEN`
- `NAME_ENDS_WITH_HYPHEN`
- `NAME_CONSECUTIVE_HYPHENS`
- `NAME_MISMATCH_DIRECTORY`

### 7.2 `description` validation (ERRORS)
Rules:
- Type: string.
- Length: **1–1024** characters.

Suggested error codes:
- `DESCRIPTION_MISSING`
- `DESCRIPTION_NOT_STRING`
- `DESCRIPTION_TOO_SHORT`
- `DESCRIPTION_TOO_LONG`

### 7.3 `compatibility` validation (ERRORS)
Rules:
- If present: type string.
- Length: **1–500** characters.

Suggested error codes:
- `COMPATIBILITY_NOT_STRING`
- `COMPATIBILITY_TOO_SHORT`
- `COMPATIBILITY_TOO_LONG`

### 7.4 `license` validation (ERRORS)
Rules:
- If present: type string (no specific license list enforcement).

Suggested error codes:
- `LICENSE_NOT_STRING`

### 7.5 `metadata` validation (ERRORS)
Rules:
- If present: must be a mapping/object.
- Keys must be strings.
- Values must be strings (reject numbers, arrays, nested objects).

Suggested error codes:
- `METADATA_NOT_OBJECT`
- `METADATA_VALUE_NOT_STRING`

### 7.6 `allowed-tools` validation (ERRORS)
Rules:
- If present: type string.
- When trimmed, must not be empty.
- Must be whitespace-delimited tokens.
- Each token must be non-empty and contain no whitespace.

Suggested error codes:
- `ALLOWED_TOOLS_NOT_STRING`
- `ALLOWED_TOOLS_EMPTY`

---

## 8) Best-practice checks (WARNINGS)

### 8.1 Markdown body warnings
- `SKILL_MD_TOO_LONG_LINES`: warn if `SKILL.md` has more than **500 lines** total.
- `SKILL_MD_MISSING_BODY`: warn if the Markdown body (after frontmatter) is empty/whitespace only.

### 8.2 Unknown top-level keys
- `UNKNOWN_TOP_LEVEL_KEY`: warn with a list of unrecognized keys in frontmatter.

### 8.3 Optional directory content warnings
If directory exists but is empty (no files):
- `SCRIPTS_DIR_EMPTY`
- `REFERENCES_DIR_EMPTY`
- `ASSETS_DIR_EMPTY`

### 8.4 File reference checks (WARNINGS)
Scan Markdown body for:
- Links: `[text](path)`
- Images: `![alt](path)`
- Also optionally scan for plain file references like `scripts/foo.sh` on their own line (simple heuristic).

For each referenced `path` that is **relative** (not starting with `http://`, `https://`, `/`, `#`):
- `REF_CONTAINS_DOTDOT`: warn if path contains `..` segments.
- `REF_TOO_DEEP`: warn if it contains more than one `/` (more than one level deep).
- `REF_MISSING_FILE`: warn if referenced file does not exist (enabled by default via `Options.CheckRefsExist=true`).
- If referenced target is a symlink escaping root and `FollowSymlinks=false`, warn or error? **Warn** by default, error only if file must exist and can’t be safely resolved.

Do not attempt to validate external URLs.

---

## 9) Reporting requirements

### 9.1 JSON output
Emit a single JSON object matching:

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

### 9.2 Text output
- Group findings under `Errors` and `Warnings` headings.
- Show `code`, `file:line` (when available), and message.
- End with a summary:
  - `X errors, Y warnings` and `VALID` / `INVALID`.

### 9.3 Ordering
- Stable ordering: sort findings by file, line, code.
- Errors before warnings.

---

## 10) Implementation details (Go)

### YAML parsing (line numbers)
Use `gopkg.in/yaml.v3` to parse into `yaml.Node` and also decode into a `map[string]any` for convenient typing.

Strategy:
- Parse YAML into a `yaml.Node` tree to find line numbers for keys (`name`, `description`, etc.).
- Decode YAML into `map[string]any` (or `map[string]interface{}`) and validate types/values.

### Filesystem and security
- Compute `skillRootAbs := filepath.Abs(path)`
- For referenced file checks:
  - Resolve: `targetAbs := filepath.Join(skillRootAbs, refPath)` then `filepath.Clean`
  - Ensure `strings.HasPrefix(targetAbs, skillRootAbs + string(os.PathSeparator))` (or equal to root for root files)
  - If outside: warn `REF_ESCAPES_ROOT` (and do not access the file).
- Symlink handling:
  - If following disabled, do not `EvalSymlinks` unless you first confirm it stays inside root.
  - Provide `--follow-symlinks` to allow fully resolving.
- Use `io/fs` (`fs.WalkDir`) for optional directory emptiness checks.

---

## 11) Test suite (mandatory)

Use Go’s standard testing (`testing` package). Add fixture directories under `testdata/`.

Create tests that cover:

1. **Valid minimal skill**
   - Directory name matches `name`
   - `SKILL.md` with correct frontmatter and short body

2. **Missing `SKILL.md`**
   - Expect error `SKILL_MD_MISSING` (choose exact code; be consistent)

3. **Frontmatter errors**
   - Missing starting `---`
   - Missing closing `---`
   - Invalid YAML
   - YAML not a mapping

4. **Name violations**
   - uppercase letter
   - starts with `-`
   - ends with `-`
   - contains `--`
   - length 0 and 65
   - mismatch with directory name

5. **Description length**
   - length 0
   - length 1025

6. **Compatibility length**
   - 0 and 501 (when present)

7. **Metadata validation**
   - metadata is list (invalid)
   - metadata has non-string values

8. **allowed-tools**
   - empty/whitespace only invalid
   - normal token list valid

9. **Reference warnings**
   - missing referenced file → `REF_MISSING_FILE`
   - deep path (e.g. `references/deep/nested.md`) → `REF_TOO_DEEP`
   - `../secret.txt` → `REF_CONTAINS_DOTDOT` (and/or `REF_ESCAPES_ROOT`)
   - existing one-level deep file should not warn

10. **Line count warning**
   - `SKILL.md` > 500 lines → warning

11. **Strict mode**
   - Warning-only fixture becomes invalid when `Strict=true`

Test pattern recommendation:
- `ValidateSkill` returns `Result` with slices; assert codes present.
- Provide helper `hasFinding(result, level, code)`.

---

## 12) README requirements

Write `README.md` including:
- What it validates
- Installation:
  - `go install .../cmd/sklint@latest`
  - or binaries from Releases (if you implement CI)
- Usage examples
- Output examples (text + json)
- Exit code meanings

---

## 13) Optional CI (recommended)

Add GitHub Actions workflow that:
- Runs `go test ./...`
- Builds binaries for:
  - linux/amd64, linux/arm64
  - darwin/amd64, darwin/arm64
  - windows/amd64
- Uploads as release artifacts on tag pushes.

---

## 14) Deliverables checklist

- [ ] `cmd/sklint/main.go` with flags and exit codes
- [ ] `pkg/validator` library with `ValidateSkill`
- [ ] JSON and text reporters
- [ ] Full test suite with fixtures in `testdata/`
- [ ] README
- [ ] (Optional) CI for releases

---

## 15) Acceptance criteria

The validator is accepted when:
- All mandatory tests pass (`go test ./...`).
- CLI returns correct exit codes.
- JSON output is stable and matches schema.
- It catches all specified errors and emits warnings as described.
- It is safe by default (does not follow symlinks outside the skill root).
