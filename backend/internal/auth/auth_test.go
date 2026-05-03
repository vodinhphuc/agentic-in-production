package auth

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestVerifyPassword_AcceptsCorrect(t *testing.T) {
	hash, err := HashPassword("hunter2")
	require.NoError(t, err)
	require.True(t, VerifyPassword(hash, "hunter2"))
	require.False(t, VerifyPassword(hash, "wrong"))
}

func TestIssueAndParseToken_RoundTrips(t *testing.T) {
	a := NewAuthenticator([]byte("0123456789abcdef0123456789abcdef"), "admin", "ignored", 1*time.Hour)
	tok, err := a.IssueToken("admin")
	require.NoError(t, err)
	claims, err := a.ParseToken(tok)
	require.NoError(t, err)
	require.Equal(t, "admin", claims.Subject)
}

func TestParseToken_RejectsTampered(t *testing.T) {
	a := NewAuthenticator([]byte("0123456789abcdef0123456789abcdef"), "admin", "ignored", 1*time.Hour)
	tok, _ := a.IssueToken("admin")
	_, err := a.ParseToken(tok + "x")
	require.Error(t, err)
}
