package audit

import (
	"context"
	"encoding/json"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/phucvd2512/agentic-in-production/backend/internal/db"
	"github.com/phucvd2512/agentic-in-production/backend/internal/sessions"
)

func TestAppendAndList(t *testing.T) {
	if os.Getenv("AIP_SKIP_DB_TESTS") == "1" {
		t.Skip()
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	d, err := db.Open(ctx)
	if err != nil {
		t.Skipf("postgres not reachable (%v)", err)
	}
	defer d.Close()

	sess, err := sessions.NewStore(d).Create(context.Background(), "admin", "mock-trino-flavored")
	require.NoError(t, err)

	a := NewLog(d)
	payload, _ := json.Marshal(map[string]any{"type": "tool_call_start", "call_id": "c1", "tool": "execute_query", "args": map[string]any{"sql": "select 1"}})
	require.NoError(t, a.Append(ctx, sess.ID, "run_1", "tool_call_start", payload))

	entries, err := a.List(ctx, sess.ID)
	require.NoError(t, err)
	require.Len(t, entries, 1)
	require.Equal(t, "run_1", entries[0].RunID)
	require.Equal(t, "tool_call_start", entries[0].Kind)
}
