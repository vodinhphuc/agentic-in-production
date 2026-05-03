package sessions

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/phucvd2512/agentic-in-production/backend/internal/db"
)

func openDB(t *testing.T) *db.DB {
	t.Helper()
	if os.Getenv("AIP_SKIP_DB_TESTS") == "1" {
		t.Skip("AIP_SKIP_DB_TESTS=1")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	d, err := db.Open(ctx)
	if err != nil {
		t.Skipf("postgres not reachable (%v)", err)
	}
	t.Cleanup(d.Close)
	return d
}

func TestCreateAndGet(t *testing.T) {
	d := openDB(t)
	ctx := context.Background()
	s := NewStore(d)

	created, err := s.Create(ctx, "admin", "mock-trino-flavored")
	require.NoError(t, err)
	require.NotEmpty(t, created.ID)

	got, err := s.Get(ctx, created.ID)
	require.NoError(t, err)
	require.Equal(t, created.ID, got.ID)
	require.Equal(t, "mock-trino-flavored", got.AgentName)
}

func TestSetPlatformConvID(t *testing.T) {
	d := openDB(t)
	ctx := context.Background()
	s := NewStore(d)

	created, _ := s.Create(ctx, "admin", "mock-trino-flavored")
	require.NoError(t, s.SetPlatformConvID(ctx, created.ID, "conv_xyz"))
	got, _ := s.Get(ctx, created.ID)
	require.NotNil(t, got.PlatformConvID)
	require.Equal(t, "conv_xyz", *got.PlatformConvID)
}
