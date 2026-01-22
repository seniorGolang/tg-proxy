package mongo

import (
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

// GetIndexModels возвращает все необходимые индексы для коллекции projects
// Уникальный индекс на поле alias (используется в GetProject, UpdateProject, DeleteProject, Exists, GetProjectToken)
// Индекс для сортировки по created_at (используется в ListProjects с ORDER BY created_at DESC)
// Индекс для поиска по source_name (может использоваться для фильтрации проектов по источнику)
// Индекс для поиска по repo_url (может использоваться для поиска проектов по URL репозитория)
func GetIndexModels() (indexModels []mongo.IndexModel) {

	indexModels = []mongo.IndexModel{
		{
			Keys:    bson.D{{Key: "alias", Value: 1}},
			Options: options.Index().SetUnique(true),
		},
		{
			Keys: bson.D{{Key: "created_at", Value: -1}},
		},
		{
			Keys: bson.D{{Key: "source_name", Value: 1}},
		},
		{
			Keys: bson.D{{Key: "repo_url", Value: 1}},
		},
	}

	return
}
