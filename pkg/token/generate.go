package token
import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/google/uuid"
)

func Generate(tokenType string, port int, note string, expiresIn time.Duration) (*Token, error) {
	id := uuid.New().String()

	secretBytes := make([]byte, 32)
	if _, err := rand.Read(secretBytes); err != nil {
		return nil, fmt.Errorf("failed to generate secret: %w", err)
	}

	secret := hex.EncodeToString(secretBytes)
	callbackURL := fmt.Sprintf("http://localhost:%d/t/%s/%s", port, id, secret)

	var expiresAt *time.Time
	if expiresIn > 0 {
		t := time.Now().UTC().Add(expiresIn)
		expiresAt = &t
	}

	return &Token{
		ID:          id,
		Type:        tokenType,
		Secret:      secret,
		CallbackURL: callbackURL,
		CreatedAt:   time.Now().UTC(),
		ExpiresAt:   expiresAt,
		Note:        note,
		Triggered:   false,
	}, nil
}