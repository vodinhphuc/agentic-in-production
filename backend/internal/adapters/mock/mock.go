package mock

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"

	"github.com/phucvd2512/agentic-in-production/backend/internal/adapters"
)

type Config struct {
	ScenarioDir string `json:"scenario_dir"`
}

type Adapter struct {
	scenarios []scenario
}

type scenario struct {
	Name   string     `yaml:"name"`
	Match  matchRule  `yaml:"match"`
	Events []rawEvent `yaml:"events"`
}

type matchRule struct {
	Any         bool     `yaml:"any"`
	ContainsAny []string `yaml:"contains_any"`
}

type rawEvent map[string]any

func New(c Config) (*Adapter, error) {
	dir := c.ScenarioDir
	if dir == "" {
		dir = "scenarios"
	}
	matches, err := filepath.Glob(filepath.Join(dir, "*.yaml"))
	if err != nil {
		return nil, err
	}
	if len(matches) == 0 {
		return nil, errors.New("no scenarios in " + dir)
	}

	var scenarios []scenario
	var hasDefault bool
	for _, m := range matches {
		b, err := os.ReadFile(m)
		if err != nil {
			return nil, err
		}
		var s scenario
		if err := yaml.Unmarshal(b, &s); err != nil {
			return nil, fmt.Errorf("%s: %w", m, err)
		}
		if s.Match.Any {
			hasDefault = true
		}
		scenarios = append(scenarios, s)
	}
	if !hasDefault {
		return nil, errors.New("at least one scenario must have match: any: true (the default)")
	}
	return &Adapter{scenarios: scenarios}, nil
}

func init() {
	adapters.Register("mock", func(configJSON []byte) (adapters.AgentPlatformAdapter, error) {
		var c Config
		if len(configJSON) > 0 {
			if err := json.Unmarshal(configJSON, &c); err != nil {
				return nil, err
			}
		}
		return New(c)
	})
}

func (a *Adapter) Name() string { return "mock" }

func (a *Adapter) Capabilities(_ context.Context) (adapters.Capabilities, error) {
	return adapters.Capabilities{Name: "mock", Tools: []string{"describe_table", "execute_query", "list_tables", "list_schemas", "list_catalogs"}}, nil
}

func (a *Adapter) StartConversation(_ context.Context, sess adapters.Session) (string, error) {
	return "mock_conv_" + sess.ID, nil
}

func (a *Adapter) EndConversation(_ context.Context, _ string) error { return nil }

func (a *Adapter) Run(ctx context.Context, req adapters.RunRequest) (<-chan adapters.AgentEvent, error) {
	s := a.pick(req.UserMessage)
	out := make(chan adapters.AgentEvent, len(s.Events))
	go func() {
		defer close(out)
		for _, raw := range s.Events {
			ev := substitute(raw, req)
			b, err := json.Marshal(ev)
			if err != nil {
				return
			}
			t, _ := ev["type"].(string)
			runID, _ := ev["run_id"].(string)
			select {
			case <-ctx.Done():
				return
			case out <- adapters.AgentEvent{Type: t, RunID: runID, Payload: b}:
			}
		}
	}()
	return out, nil
}

func (a *Adapter) pick(userMsg string) scenario {
	lower := strings.ToLower(userMsg)
	var def *scenario
	for i := range a.scenarios {
		s := &a.scenarios[i]
		if s.Match.Any {
			def = s
			continue
		}
		for _, kw := range s.Match.ContainsAny {
			if strings.Contains(lower, strings.ToLower(kw)) {
				return *s
			}
		}
	}
	return *def
}

// substitute replaces __RUN_ID__ and __USER_MESSAGE__ tokens in event values.
func substitute(in rawEvent, req adapters.RunRequest) rawEvent {
	out := make(rawEvent, len(in))
	for k, v := range in {
		out[k] = sub(v, req)
	}
	return out
}

func sub(v any, req adapters.RunRequest) any {
	switch t := v.(type) {
	case string:
		s := strings.ReplaceAll(t, "__RUN_ID__", req.RunID)
		s = strings.ReplaceAll(s, "__USER_MESSAGE__", req.UserMessage)
		return s
	case map[string]any:
		o := make(map[string]any, len(t))
		for k, vv := range t {
			o[k] = sub(vv, req)
		}
		return o
	case []any:
		o := make([]any, len(t))
		for i, vv := range t {
			o[i] = sub(vv, req)
		}
		return o
	default:
		return v
	}
}
