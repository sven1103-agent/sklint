package validator

type Options struct {
	Strict         bool
	NoWarn         bool
	FollowSymlinks bool
	CheckRefsExist bool
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
