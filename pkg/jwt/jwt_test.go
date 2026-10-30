package jwt

import (
	"errors"
	"testing"
	"time"

	jwtlib "github.com/golang-jwt/jwt/v5"
)

func TestGenerateToken_RoundTrip(t *testing.T) {
	m := New("top-secret", time.Hour)

	token, err := m.GenerateToken("42")
	if err != nil {
		t.Fatalf("GenerateToken() error = %v", err)
	}
	if token == "" {
		t.Fatal("GenerateToken() returned empty token")
	}

	sub, err := m.ParseToken(token)
	if err != nil {
		t.Fatalf("ParseToken() error = %v", err)
	}
	if sub != "42" {
		t.Errorf("ParseToken() subject = %q, want %q", sub, "42")
	}
}

func TestParseToken_InvalidToken(t *testing.T) {
	m := New("top-secret", time.Hour)

	if _, err := m.ParseToken("not-a-jwt"); err == nil {
		t.Fatal("ParseToken() expected error for malformed token")
	}
}

func TestParseToken_WrongSecret(t *testing.T) {
	token, err := New("first-secret", time.Hour).GenerateToken("1")
	if err != nil {
		t.Fatalf("GenerateToken() error = %v", err)
	}

	if _, err := New("other-secret", time.Hour).ParseToken(token); err == nil {
		t.Fatal("ParseToken() expected error for token signed with another secret")
	}
}

func TestParseToken_Expired(t *testing.T) {
	token, err := New("top-secret", -time.Minute).GenerateToken("1")
	if err != nil {
		t.Fatalf("GenerateToken() error = %v", err)
	}

	if _, err := New("top-secret", time.Hour).ParseToken(token); err == nil {
		t.Fatal("ParseToken() expected error for expired token")
	}
}

func TestParseToken_UnexpectedSigningMethod(t *testing.T) {
	unsigned, err := jwtlib.NewWithClaims(jwtlib.SigningMethodNone, jwtlib.RegisteredClaims{
		Subject:   "1",
		ExpiresAt: jwtlib.NewNumericDate(time.Now().Add(time.Hour)),
	}).SignedString(jwtlib.UnsafeAllowNoneSignatureType)
	if err != nil {
		t.Fatalf("failed to build unsigned token: %v", err)
	}

	_, err = New("top-secret", time.Hour).ParseToken(unsigned)
	if !errors.Is(err, ErrUnexpectedSigningMethod) {
		t.Fatalf("ParseToken() error = %v, want ErrUnexpectedSigningMethod", err)
	}
}
