package mock

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/phucvd2512/agentic-in-production/backend/internal/adapters"
	"github.com/phucvd2512/agentic-in-production/backend/internal/protocol"
)

func TestRun_DefaultScenario_EchoesMessage(t *testing.T) {
	a, err := New(Config{ScenarioDir: "scenarios"})
	require.NoError(t, err)

	req := adapters.RunRequest{
		RunID:       "run_1",
		Session:     adapters.Session{ID: "s1", UserID: "admin", AgentName: "mock-trino-flavored"},
		UserMessage: "hello there",
	}
	ch, err := a.Run(context.Background(), req)
	require.NoError(t, err)

	var got []map[string]any
	for ev := range ch {
		var m map[string]any
		require.NoError(t, json.Unmarshal(ev.Payload, &m))
		got = append(got, m)
	}
	require.NotEmpty(t, got)
	require.Equal(t, "run_started", got[0]["type"])
	require.Equal(t, "run_finished", got[len(got)-1]["type"])

	// At least one text_delta should contain the user message verbatim.
	var saw bool
	for _, m := range got {
		if m["type"] == "text_delta" && strings.Contains(m["text"].(string), "hello there") {
			saw = true
			break
		}
	}
	require.True(t, saw, "expected user message to appear in echoed text")
}

func TestRun_TrinoInvestigationScenario_TriggersOnKeyword(t *testing.T) {
	a, err := New(Config{ScenarioDir: "scenarios"})
	require.NoError(t, err)

	req := adapters.RunRequest{
		RunID:       "run_2",
		Session:     adapters.Session{ID: "s1", UserID: "admin", AgentName: "mock-trino-flavored"},
		UserMessage: "investigate WIN-WS-014",
	}
	ch, _ := a.Run(context.Background(), req)
	var types []string
	for ev := range ch {
		var m map[string]any
		_ = json.Unmarshal(ev.Payload, &m)
		types = append(types, m["type"].(string))
	}
	// expect at least one tool_call_start in the trino scenario
	var sawTool bool
	for _, t := range types {
		if t == "tool_call_start" {
			sawTool = true
			break
		}
	}
	require.True(t, sawTool, "trino-investigation scenario should emit tool calls")
}

func TestRun_AllEmittedEventsAreSchemaValid(t *testing.T) {
	a, err := New(Config{ScenarioDir: "scenarios"})
	require.NoError(t, err)

	for _, msg := range []string{"hello", "investigate WIN-WS-014", "show me powershell"} {
		req := adapters.RunRequest{RunID: "r", Session: adapters.Session{ID: "s"}, UserMessage: msg}
		ch, _ := a.Run(context.Background(), req)
		for ev := range ch {
			require.NoError(t, protocol.ValidateEvent(ev.Payload), "schema violation for: %s", string(ev.Payload))
		}
	}
}
