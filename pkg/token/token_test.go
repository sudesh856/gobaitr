package token

import (
	"strings"
	"testing"
)

func TestGenerateToken(t *testing.T) {
	tok, err := Generate("url", 8080, "test", 0)
	if err != nil {
		t.Fatal(err)
	}
	if len(tok.ID) != 36 {
		t.Errorf("expected UUID length 36, got %d", len(tok.ID))
	}

	if len(tok.Secret) != 64 {
		t.Errorf("expected secret length 64, got %d", len(tok.Secret))
	}

	// Callback URL must contain ID and secret
	if !strings.Contains(tok.CallbackURL, tok.ID) {
		t.Error("callback URL must contain token ID")
	}
	if !strings.Contains(tok.CallbackURL, tok.Secret) {
		t.Error("callback URL must contain secret")
	}
}

func TestTokenSecretEntropy(t *testing.T) {
	secrets := make(map[string]bool)
	for i := 0; i < 1000; i++ {
		tok, err := Generate("url", 8080, "", 0)
		if err != nil {
			t.Fatal(err)
		}
		if secrets[tok.Secret] {
			t.Fatalf("duplicate secret found at iteration %d", i)
		}
		secrets[tok.Secret] = true
	}
}