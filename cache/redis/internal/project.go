package internal

import (
	"time"

	"github.com/seniorGolang/tg-proxy/model/domain"
)

// Project представляет документ проекта в Redis
type Project struct {
	Alias          string    `json:"alias"`
	RepoURL        string    `json:"repo_url"`
	EncryptedToken string    `json:"encrypted_token"`
	Token          string    `json:"token"`
	Description    string    `json:"description"`
	SourceName     string    `json:"source_name"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

// ToDomain преобразует документ в доменную модель
func (d Project) ToDomain() (project domain.Project) {

	return domain.Project{
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

// FromDomain создает документ из доменной модели
func FromDomain(project domain.Project) (doc Project) {

	return Project{
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
