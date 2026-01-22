package mongo

import (
	"github.com/seniorGolang/tg-proxy/model/domain"
	"github.com/seniorGolang/tg-proxy/storage/mongo/internal"
)

// toDocument преобразует доменную модель в документ MongoDB
func toDocument(project domain.Project) (doc internal.ProjectDocument) {
	return internal.ProjectDocument{
		Alias:          project.Alias,
		RepoURL:        project.RepoURL,
		EncryptedToken: project.EncryptedToken,
		Description:    project.Description,
		SourceName:     project.SourceName,
		CreatedAt:      project.CreatedAt,
		UpdatedAt:      project.UpdatedAt,
	}
}

// toDomain преобразует документ MongoDB в доменную модель
func toDomain(doc internal.ProjectDocument) (project domain.Project) {
	return domain.Project{
		Alias:          doc.Alias,
		RepoURL:        doc.RepoURL,
		EncryptedToken: doc.EncryptedToken,
		Description:    doc.Description,
		SourceName:     doc.SourceName,
		CreatedAt:      doc.CreatedAt,
		UpdatedAt:      doc.UpdatedAt,
	}
}

// toUpdateDocument преобразует доменную модель в документ для обновления MongoDB
func toUpdateDocument(project domain.Project) (doc internal.ProjectUpdateDocument) {
	return internal.ProjectUpdateDocument{
		RepoURL:        project.RepoURL,
		EncryptedToken: project.EncryptedToken,
		Description:    project.Description,
		SourceName:     project.SourceName,
		UpdatedAt:      project.UpdatedAt,
	}
}
