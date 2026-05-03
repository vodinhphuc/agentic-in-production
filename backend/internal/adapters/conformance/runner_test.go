package conformance

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/phucvd2512/agentic-in-production/backend/internal/adapters/mock"
)

func TestMockAdapter_PassesAllScenarios(t *testing.T) {
	scenarios, err := LoadScenarios("scenarios")
	require.NoError(t, err)
	require.NotEmpty(t, scenarios)

	ad, err := mock.New(mock.Config{ScenarioDir: "../mock/scenarios"})
	require.NoError(t, err)

	for _, s := range scenarios {
		s := s
		t.Run(s.Name, func(t *testing.T) {
			got, err := Run(context.Background(), ad, s)
			require.NoError(t, err)
			require.NoError(t, Verify(s, got), "got events: %v", got)
		})
	}
}
