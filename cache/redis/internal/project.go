package internal

import (
	"time"

	"github.com/google/uuid"

	"github.com/seniorGolang/tg-proxy/model/domain"
)

type Project struct {
	ID             uuid.UUID `json:"id"`
	Alias          string    `json:"alias"`
	RepoURL        string    `json:"repo_url"`
	EncryptedToken string    `json:"encrypted_token"`
	Token          string    `json:"token"`
	Description    string    `json:"description"`
	SourceName     string    `json:"source_name"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

func (d Project) ToDomain() (project domain.Project) {

	return domain.Project{
		ID:             d.ID,
		Alias:          d.Alias,
		RepoURL:        d.RepoURL,
		EncryptedToken: d.EncryptedToken,
		Token:          d.Token,
		Description:    d.Description,
		SourceName:     d.SourceName,
		CreatedAt:      d.CreatedAt,
		UpdatedAt:      d.UpdatedAt,
	}
}

func FromDomain(project domain.Project) (doc Project) {

	return Project{
		ID:             project.ID,
		Alias:          project.Alias,
		RepoURL:        project.RepoURL,
		EncryptedToken: project.EncryptedToken,
		Token:          project.Token,
		Description:    project.Description,
		SourceName:     project.SourceName,
		CreatedAt:      project.CreatedAt,
		UpdatedAt:      project.UpdatedAt,
	}
}
