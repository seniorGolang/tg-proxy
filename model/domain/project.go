package domain

import "time"

type Project struct {
	Alias          string
	RepoURL        string
	EncryptedToken string
	Token          string
	Description    string
	SourceName     string
	CreatedAt      time.Time
	UpdatedAt      time.Time
}
