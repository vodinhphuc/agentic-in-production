package db

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

// Skip when Postgres isn't reachable so unit tests still run on a bare workstation.
func TestOpen_PingsLivePostgres(t *testing.T) {
	if os.Getenv("AIP_SKIP_DB_TESTS") == "1" {
		t.Skip("AIP_SKIP_DB_TESTS=1")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	d, err := Open(ctx)
	if err != nil {
		t.Skipf("postgres not reachable; skipping (%v)", err)
	}
	defer d.Close()
	require.NoError(t, d.Pool.Ping(ctx))
}
