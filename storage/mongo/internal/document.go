package internal

import (
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

// ProjectDocument представляет документ проекта в MongoDB
type ProjectDocument struct {
	ID             bson.ObjectID `bson:"_id,omitempty"`
	Alias          string        `bson:"alias"`
	RepoURL        string        `bson:"repo_url"`
	EncryptedToken string        `bson:"encrypted_token,omitempty"`
	Description    string        `bson:"description,omitempty"`
	SourceName     string        `bson:"source_name,omitempty"`
	CreatedAt      time.Time     `bson:"created_at"`
	UpdatedAt      time.Time     `bson:"updated_at"`
}

// ProjectUpdateDocument представляет документ для обновления проекта в MongoDB
type ProjectUpdateDocument struct {
	RepoURL        string    `bson:"repo_url"`
	EncryptedToken string    `bson:"encrypted_token,omitempty"`
	Description    string    `bson:"description,omitempty"`
	SourceName     string    `bson:"source_name,omitempty"`
	UpdatedAt      time.Time `bson:"updated_at"`
}
