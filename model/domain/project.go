package domain

import (
	"time"

	"github.com/google/uuid"
)

type Project struct {
	ID             uuid.UUID
	Alias          string
	RepoURL        string
	EncryptedToken string
	Token          string
	Description    string
	SourceName     string
	CreatedAt      time.Time
	UpdatedAt      time.Time
}
