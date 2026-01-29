package mongo

import (
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

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
