package mongo

import (
	"context"
	"fmt"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"

	"github.com/seniorGolang/tg-proxy/storage/mongo/internal"
)

func (s *Storage) ensureCatalogVersion(ctx context.Context) (err error) {

	filter := bson.M{"_id": s.catalogVersionDocID}
	update := bson.M{
		"$setOnInsert": bson.M{"major": 1, "minor": 0, "patch": 0},
	}
	opts := options.UpdateOne().SetUpsert(true)

	_, err = s.versionCollection.UpdateOne(ctx, filter, update, opts)
	return
}

type bumpKind int

const (
	bumpMajor bumpKind = iota
	bumpMinor
	bumpPatch
)

func (s *Storage) GetCatalogVersion(ctx context.Context) (version string, err error) {

	var doc internal.CatalogVersionDocument
	if err = s.versionCollection.FindOne(ctx, bson.M{"_id": s.catalogVersionDocID}).Decode(&doc); err != nil {
		if err == mongo.ErrNoDocuments {
			return "1.0.0", nil
		}
		return
	}

	version = fmt.Sprintf("%d.%d.%d", doc.Major, doc.Minor, doc.Patch)
	return
}

func (s *Storage) bumpCatalogVersion(ctx context.Context, kind bumpKind) (err error) {

	filter := bson.M{"_id": s.catalogVersionDocID}
	opts := options.FindOneAndUpdate().SetUpsert(true).SetReturnDocument(options.After)

	var update bson.M
	switch kind {
	case bumpMajor:
		update = bson.M{
			"$setOnInsert": bson.M{"major": 1},
			"$inc":         bson.M{"major": 1},
			"$set":         bson.M{"minor": 0, "patch": 0},
		}
	case bumpMinor:
		update = bson.M{
			"$setOnInsert": bson.M{"major": 1},
			"$inc":         bson.M{"minor": 1},
			"$set":         bson.M{"patch": 0},
		}
	case bumpPatch:
		update = bson.M{
			"$setOnInsert": bson.M{"major": 1, "minor": 0},
			"$inc":         bson.M{"patch": 1},
		}
	}

	var doc internal.CatalogVersionDocument
	if err = s.versionCollection.FindOneAndUpdate(ctx, filter, update, opts).Decode(&doc); err != nil {
		return
	}

	return
}
