package auth

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

type Claims struct {
	Subject string `json:"sub"`
	jwt.RegisteredClaims
}

type Authenticator struct {
	signingKey  []byte
	adminUser   string
	adminPwHash string
	tokenTTL    time.Duration
}

func NewAuthenticator(key []byte, adminUser, adminPwHash string, ttl time.Duration) *Authenticator {
	return &Authenticator{signingKey: key, adminUser: adminUser, adminPwHash: adminPwHash, tokenTTL: ttl}
}

func HashPassword(plain string) (string, error) {
	b, err := bcrypt.GenerateFromPassword([]byte(plain), bcrypt.DefaultCost)
	return string(b), err
}

func VerifyPassword(hash, plain string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(plain)) == nil
}

func (a *Authenticator) Login(user, pw string) (string, error) {
	if user != a.adminUser || !VerifyPassword(a.adminPwHash, pw) {
		return "", errors.New("bad credentials")
	}
	return a.IssueToken(user)
}

func (a *Authenticator) IssueToken(sub string) (string, error) {
	t := jwt.NewWithClaims(jwt.SigningMethodHS256, Claims{
		Subject: sub,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(a.tokenTTL)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	})
	return t.SignedString(a.signingKey)
}

func (a *Authenticator) ParseToken(raw string) (*Claims, error) {
	tok, err := jwt.ParseWithClaims(raw, &Claims{}, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return a.signingKey, nil
	})
	if err != nil {
		return nil, err
	}
	c, ok := tok.Claims.(*Claims)
	if !ok || !tok.Valid {
		return nil, errors.New("invalid token")
	}
	return c, nil
}
