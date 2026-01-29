package dto

import (
	"time"

	"github.com/google/uuid"

	"github.com/seniorGolang/tg-proxy/model/domain"
)

type ProjectCreateRequest struct {
	Alias       string `json:"alias" validate:"required,min=1,max=255"`
	RepoURL     string `json:"repo_url" validate:"required,url"`
	Token       string `json:"token,omitempty" validate:"omitempty"`
	Description string `json:"description,omitempty" validate:"omitempty,max=1000"`
	SourceName  string `json:"source_name" validate:"required"`
}

type ProjectUpdateRequest struct {
	RepoURL     *string `json:"repo_url,omitempty" validate:"omitempty,url"`
	Token       *string `json:"token,omitempty" validate:"omitempty"`
	Description *string `json:"description,omitempty" validate:"omitempty,max=1000"`
	SourceName  *string `json:"source_name,omitempty" validate:"omitempty,required"`
}

type ProjectResponse struct {
	ID          uuid.UUID `json:"id"`
	Alias       string    `json:"alias"`
	RepoURL     string    `json:"repo_url"`
	Description string    `json:"description,omitempty"`
	SourceName  string    `json:"source_name,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

func (dto *ProjectCreateRequest) ToDomain() (project domain.Project) {
	return domain.Project{
		Alias:       dto.Alias,
		RepoURL:     dto.RepoURL,
		Token:       dto.Token,
		Description: dto.Description,
		SourceName:  dto.SourceName,
	}
}

func (dto *ProjectUpdateRequest) ToDomain(alias string) (project domain.Project) {

	project.Alias = alias

	if dto.RepoURL != nil {
		project.RepoURL = *dto.RepoURL
	}
	if dto.Token != nil {
		project.Token = *dto.Token
	}
	if dto.Description != nil {
		project.Description = *dto.Description
	}
	if dto.SourceName != nil {
		project.SourceName = *dto.SourceName
	}

	return project
}

func FromDomain(project domain.Project) (resp ProjectResponse) {
	return ProjectResponse{
		ID:          project.ID,
		Alias:       project.Alias,
		RepoURL:     project.RepoURL,
		Description: project.Description,
		SourceName:  project.SourceName,
		CreatedAt:   project.CreatedAt,
		UpdatedAt:   project.UpdatedAt,
	}
}
