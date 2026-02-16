package report

import (
	"encoding/json"
	"testing"

	"sklint/pkg/validator"
)

func TestRenderJSON(t *testing.T) {
	result := validator.Result{
		Path:  "/tmp/skill",
		Valid: true,
	}
	out, err := RenderJSON(result)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	var decoded validator.Result
	if err := json.Unmarshal(out, &decoded); err != nil {
		t.Fatalf("unexpected json error: %v", err)
	}
	if decoded.Path != result.Path || decoded.Valid != result.Valid {
		t.Fatalf("unexpected decoded result: %#v", decoded)
	}
}
