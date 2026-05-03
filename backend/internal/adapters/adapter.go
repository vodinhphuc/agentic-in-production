package adapters

import (
	"context"
	"encoding/json"
)

// AgentEvent is one canonical event yielded by an adapter. Payload is the raw
// JSON object conforming to protocols/agent-events.schema.json.
type AgentEvent struct {
	Type    string          // "run_started" | "text_delta" | ...
	RunID   string          // copied from run_started; available for downstream correlation
	Payload json.RawMessage // the full event JSON, schema-valid
}

type Capabilities struct {
	Name  string   `json:"name"`
	Model string   `json:"model,omitempty"`
	Tools []string `json:"tools,omitempty"`
}

type Session struct {
	ID        string
	UserID    string
	AgentName string
}

type RunRequest struct {
	RunID          string
	Session        Session
	PlatformConvID string // empty for first message; adapters may create one
	UserMessage    string
	ToolAllowlist  []string // backend-enforced; adapter must honor
}

type AgentPlatformAdapter interface {
	Name() string
	Capabilities(ctx context.Context) (Capabilities, error)

	// StartConversation may be a no-op (Mock) or a remote create call (real platforms).
	StartConversation(ctx context.Context, sess Session) (platformConvID string, err error)
	EndConversation(ctx context.Context, platformConvID string) error

	// Run yields canonical AgentEvents on the returned channel and closes it on completion.
	// The channel must NOT be returned in an error state; errors are emitted as Error events
	// followed by RunFinished{reason:"error"}.
	Run(ctx context.Context, req RunRequest) (<-chan AgentEvent, error)
}
