package httpapi

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/phucvd2512/agentic-in-production/backend/internal/auth"
)

func TestLogin_GoodCredentials_SetsCookie(t *testing.T) {
	hash, _ := auth.HashPassword("hunter2")
	a := auth.NewAuthenticator([]byte("0123456789abcdef0123456789abcdef"), "admin", hash, time.Hour)
	h := AuthHandler{Auth: a}

	req := httptest.NewRequest(http.MethodPost, "/api/auth/login",
		bytes.NewReader([]byte(`{"username":"admin","password":"hunter2"}`)))
	w := httptest.NewRecorder()
	h.Login(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	cookies := w.Result().Cookies()
	require.Len(t, cookies, 1)
	require.Equal(t, "aip_session", cookies[0].Name)
}

func TestLogin_BadCredentials_401(t *testing.T) {
	hash, _ := auth.HashPassword("hunter2")
	a := auth.NewAuthenticator([]byte("0123456789abcdef0123456789abcdef"), "admin", hash, time.Hour)
	h := AuthHandler{Auth: a}

	req := httptest.NewRequest(http.MethodPost, "/api/auth/login",
		bytes.NewReader([]byte(`{"username":"admin","password":"wrong"}`)))
	w := httptest.NewRecorder()
	h.Login(w, req)
	require.Equal(t, http.StatusUnauthorized, w.Code)
}
