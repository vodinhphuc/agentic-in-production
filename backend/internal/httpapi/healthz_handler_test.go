package httpapi

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestHealthz_ReturnsOK(t *testing.T) {
	r := httptest.NewRequest(http.MethodGet, "/api/healthz", nil)
	w := httptest.NewRecorder()
	HealthzHandler{Version: "test"}.ServeHTTP(w, r)
	require.Equal(t, http.StatusOK, w.Code)
	var body struct {
		Ok      bool   `json:"ok"`
		Version string `json:"version"`
	}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
	require.True(t, body.Ok)
	require.Equal(t, "test", body.Version)
}
