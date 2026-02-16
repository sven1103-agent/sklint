package report

import (
	"encoding/json"

	"github.com/sven1103-agent/sklint/pkg/validator"
)

func RenderJSON(result validator.Result) ([]byte, error) {
	return json.MarshalIndent(result, "", "  ")
}
