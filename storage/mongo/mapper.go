package mongo

import (
	"github.com/seniorGolang/tg-proxy/model/domain"
	"github.com/seniorGolang/tg-proxy/storage/mongo/internal"
)

func toDocument(project domain.Project) (doc internal.ProjectDocument) {
	return internal.ProjectDocument{
		ID:             project.ID,
		Alias:          project.Alias,
		RepoURL:        project.RepoURL,
		EncryptedToken: project.EncryptedToken,
		Description:    project.Description,
		SourceName:     project.SourceName,
		CreatedAt:      project.CreatedAt,
		UpdatedAt:      project.UpdatedAt,
	}
}

func toDomain(doc internal.ProjectDocument) (project domain.Project) {
	return domain.Project{
		ID:             doc.ID,
		Alias:          doc.Alias,
		RepoURL:        doc.RepoURL,
		EncryptedToken: doc.EncryptedToken,
		Description:    doc.Description,
		SourceName:     doc.SourceName,
		CreatedAt:      doc.CreatedAt,
		UpdatedAt:      doc.UpdatedAt,
	}
}

func toUpdateDocument(project domain.Project) (doc internal.ProjectUpdateDocument) {
	return internal.ProjectUpdateDocument{
		RepoURL:        project.RepoURL,
		EncryptedToken: project.EncryptedToken,
		Description:    project.Description,
		SourceName:     project.SourceName,
		UpdatedAt:      project.UpdatedAt,
	}
}
