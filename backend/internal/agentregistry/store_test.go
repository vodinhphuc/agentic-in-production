package agentregistry

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/phucvd2512/agentic-in-production/backend/internal/db"
)

func TestListEnabled_FindsSeed(t *testing.T) {
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

	agents, err := NewStore(d).ListEnabled(ctx)
	require.NoError(t, err)
	require.NotEmpty(t, agents)

	var found bool
	for _, a := range agents {
		if a.Name == "mock-trino-flavored" {
			found = true
			require.Equal(t, "mock", a.AdapterKind)
			break
		}
	}
	require.True(t, found, "expected mock-trino-flavored agent in registry")
}
