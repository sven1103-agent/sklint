package report

import (
	"encoding/json"

	"sklint/pkg/validator"
)

func RenderJSON(result validator.Result) ([]byte, error) {
	return json.MarshalIndent(result, "", "  ")
}
