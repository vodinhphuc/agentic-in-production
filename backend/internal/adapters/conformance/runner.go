package conformance

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"

	"github.com/phucvd2512/agentic-in-production/backend/internal/adapters"
	"github.com/phucvd2512/agentic-in-production/backend/internal/protocol"
)

type Scenario struct {
	Name                string   `yaml:"name"`
	Input               string   `yaml:"input"`
	ExpectTypes         []string `yaml:"expect_types"`
	ExpectTypesContains []string `yaml:"expect_types_contains"`
	ExpectStartsWith    string   `yaml:"expect_starts_with"`
	ExpectEndsWith      string   `yaml:"expect_ends_with"`
	ExpectMinEvents     int      `yaml:"expect_min_events"`
}

func LoadScenarios(dir string) ([]Scenario, error) {
	matches, err := filepath.Glob(filepath.Join(dir, "*.yaml"))
	if err != nil {
		return nil, err
	}
	var out []Scenario
	for _, m := range matches {
		b, err := os.ReadFile(m)
		if err != nil {
			return nil, err
		}
		var s Scenario
		if err := yaml.Unmarshal(b, &s); err != nil {
			return nil, fmt.Errorf("%s: %w", m, err)
		}
		out = append(out, s)
	}
	return out, nil
}

// Run executes a single scenario against the given adapter and returns
// the collected event types (already schema-validated) or an error
// describing the first invariant violation.
func Run(ctx context.Context, ad adapters.AgentPlatformAdapter, s Scenario) ([]string, error) {
	convID, err := ad.StartConversation(ctx, adapters.Session{ID: "conf_sess", UserID: "conf"})
	if err != nil {
		return nil, fmt.Errorf("start_conv: %w", err)
	}
	defer ad.EndConversation(ctx, convID)

	ch, err := ad.Run(ctx, adapters.RunRequest{
		RunID:          "conf_run",
		PlatformConvID: convID,
		Session:        adapters.Session{ID: "conf_sess", UserID: "conf"},
		UserMessage:    s.Input,
	})
	if err != nil {
		return nil, err
	}

	var got []string
	for ev := range ch {
		if err := protocol.ValidateEvent(ev.Payload); err != nil {
			return got, fmt.Errorf("schema: %w (raw=%s)", err, string(ev.Payload))
		}
		var m map[string]any
		_ = json.Unmarshal(ev.Payload, &m)
		got = append(got, m["type"].(string))
	}
	return got, nil
}

func Verify(s Scenario, got []string) error {
	if s.ExpectMinEvents > 0 && len(got) < s.ExpectMinEvents {
		return fmt.Errorf("expected at least %d events, got %d", s.ExpectMinEvents, len(got))
	}
	if s.ExpectStartsWith != "" && (len(got) == 0 || got[0] != s.ExpectStartsWith) {
		return fmt.Errorf("expected first event %q, got %v", s.ExpectStartsWith, got)
	}
	if s.ExpectEndsWith != "" && (len(got) == 0 || got[len(got)-1] != s.ExpectEndsWith) {
		return fmt.Errorf("expected last event %q, got %v", s.ExpectEndsWith, got)
	}
	for _, t := range s.ExpectTypesContains {
		if !contains(got, t) {
			return fmt.Errorf("expected to contain %q, got %v", t, got)
		}
	}
	if len(s.ExpectTypes) > 0 {
		// strict in-order subset check
		i := 0
		for _, want := range s.ExpectTypes {
			for i < len(got) && got[i] != want {
				i++
			}
			if i >= len(got) {
				return fmt.Errorf("expected sequence %v missing %q after position %d", s.ExpectTypes, want, i)
			}
			i++
		}
	}
	return nil
}

func contains(xs []string, x string) bool {
	for _, v := range xs {
		if v == x {
			return true
		}
	}
	return false
}
