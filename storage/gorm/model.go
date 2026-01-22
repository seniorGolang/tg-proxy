package gorm

import (
	"time"

	"gorm.io/gorm"

	"github.com/seniorGolang/tg-proxy/model/domain"
	"github.com/seniorGolang/tg-proxy/storage/gorm/generated"
)

// Project представляет модель проекта в базе данных
type Project struct {
	Alias          string    `gorm:"primaryKey;column:alias"`
	RepoURL        string    `gorm:"column:repo_url;not null;index:idx_projects_repo_url"`
	EncryptedToken string    `gorm:"column:encrypted_token"`
	Description    string    `gorm:"column:description"`
	SourceName     string    `gorm:"column:source_name;index:idx_projects_source_name"`
	CreatedAt      time.Time `gorm:"column:created_at;not null;index:idx_projects_created_at,sort:desc"`
	UpdatedAt      time.Time `gorm:"column:updated_at;not null"`
}

// TableName возвращает имя таблицы для модели Project
func (Project) TableName() string {
	return TableProjects
}

// ToDomain преобразует модель в доменную модель
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

// FromDomain создает модель из доменной модели
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

// BeforeCreate хук перед созданием записи
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

// BeforeUpdate хук перед обновлением записи
func (p *Project) BeforeUpdate(tx *gorm.DB) (err error) {
	p.UpdatedAt = time.Now()
	return
}

// GetOmitFields возвращает имена колонок полей, которые нужно исключить при обновлении
// Использует сгенерированные GORM CLI field helpers для получения имен колонок
// Включает первичный ключ и поля с автогенерацией
func GetOmitFields() (fields []string) {

	// Используем Column() метод из сгенерированных field helpers для получения имен колонок
	fields = []string{
		generated.Project.Alias.Column().Name,     // Первичный ключ
		generated.Project.CreatedAt.Column().Name, // Поле с автогенерацией
	}

	return
}
