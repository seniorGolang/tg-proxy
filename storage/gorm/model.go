package gorm

import (
	"time"

	"gorm.io/gorm"

	"github.com/seniorGolang/tg-proxy/model/domain"
	"github.com/seniorGolang/tg-proxy/storage/gorm/generated"
)

type Project struct {
	Alias          string    `gorm:"primaryKey;column:alias"`
	RepoURL        string    `gorm:"column:repo_url;not null;index:idx_projects_repo_url"`
	EncryptedToken string    `gorm:"column:encrypted_token"`
	Description    string    `gorm:"column:description"`
	SourceName     string    `gorm:"column:source_name;index:idx_projects_source_name"`
	CreatedAt      time.Time `gorm:"column:created_at;not null;index:idx_projects_created_at,sort:desc"`
	UpdatedAt      time.Time `gorm:"column:updated_at;not null"`
}

func (Project) TableName() string {
	return TableProjects
}

func (p Project) ToDomain() (project domain.Project) {
	return domain.Project{
		Alias:          p.Alias,
		RepoURL:        p.RepoURL,
		EncryptedToken: p.EncryptedToken,
		Description:    p.Description,
		SourceName:     p.SourceName,
		CreatedAt:      p.CreatedAt,
		UpdatedAt:      p.UpdatedAt,
	}
}

func FromDomain(project domain.Project) (p Project) {
	return Project{
		Alias:          project.Alias,
		RepoURL:        project.RepoURL,
		EncryptedToken: project.EncryptedToken,
		Description:    project.Description,
		SourceName:     project.SourceName,
		CreatedAt:      project.CreatedAt,
		UpdatedAt:      project.UpdatedAt,
	}
}

func (p *Project) BeforeCreate(tx *gorm.DB) (err error) {
	now := time.Now()
	if p.CreatedAt.IsZero() {
		p.CreatedAt = now
	}
	if p.UpdatedAt.IsZero() {
		p.UpdatedAt = now
	}
	return
}

func (p *Project) BeforeUpdate(tx *gorm.DB) (err error) {
	p.UpdatedAt = time.Now()
	return
}

func GetOmitFields() (fields []string) {

	fields = []string{
		generated.Project.Alias.Column().Name,
		generated.Project.CreatedAt.Column().Name,
	}

	return
}

type CatalogVersion struct {
	ID    int `gorm:"primaryKey;column:id"`
	Major int `gorm:"column:major;not null"`
	Minor int `gorm:"column:minor;not null"`
	Patch int `gorm:"column:patch;not null"`
}

func (CatalogVersion) TableName() string {
	return TableCatalogVersion
}
