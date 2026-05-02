package protocol

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestValidateEvent_AcceptsAllExamples(t *testing.T) {
	matches, err := filepath.Glob("../../../protocols/examples/*.json")
	require.NoError(t, err)
	require.NotEmpty(t, matches, "no example files found")

	for _, path := range matches {
		path := path
		t.Run(filepath.Base(path), func(t *testing.T) {
			b, err := os.ReadFile(path)
			require.NoError(t, err)

			var events []json.RawMessage
			require.NoError(t, json.Unmarshal(b, &events))
			require.NotEmpty(t, events)

			for i, ev := range events {
				if err := ValidateEvent(ev); err != nil {
					t.Fatalf("event %d failed: %v", i, err)
				}
			}
		})
	}
}

func TestValidateEvent_RejectsMalformed(t *testing.T) {
	cases := map[string]string{
		"unknown_type":      `{"type":"flibbertigibbet"}`,
		"missing_run_id":    `{"type":"run_started"}`,
		"unpaired_call_id":  `{"type":"tool_call_end","ok":true}`,
		"bad_finish_reason": `{"type":"run_finished","reason":"banana"}`,
		"extra_property":    `{"type":"text_delta","text":"hi","extra":1}`,
	}
	for name, raw := range cases {
		t.Run(name, func(t *testing.T) {
			require.Error(t, ValidateEvent([]byte(raw)))
		})
	}
}
