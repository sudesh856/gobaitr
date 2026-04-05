package token
import (
	"time"
)

type Token struct {
	ID string
	Type string 
	Secret string
	CallbackURL string
	CreatedAt time.Time
	ExpiresAt   *time.Time
	Note string
	Triggered bool
	TriggeredAt *time.Time
	TriggeredBy string
}

