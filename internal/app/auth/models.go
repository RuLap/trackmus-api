package auth

import uuid "github.com/google/uuid"

type Provider string

const (
	LocalProvider  Provider = "local"
	GoogleProvider Provider = "google"
)

func (p Provider) IsValid() bool {
	return p == LocalProvider || p == GoogleProvider
}

type User struct {
	ID             uuid.UUID `db:"id"`
	Email          string    `db:"email"`
	Provider       Provider  `db:"provider"`
	ProviderID     *string   `db:"provider_id"`
	EmailConfirmed bool      `db:"email_confirmed"`
	Password       *string   `db:"password"`
}
