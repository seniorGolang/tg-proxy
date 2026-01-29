package mongo

import (
	"context"
	"time"

	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"

	"github.com/seniorGolang/tg-proxy/model/domain"
	"github.com/seniorGolang/tg-proxy/storage"
	"github.com/seniorGolang/tg-proxy/storage/mongo/internal"
)

func (s *Storage) GetProject(ctx context.Context, alias string) (project domain.Project, found bool, err error) {

	var doc internal.ProjectDocument
	if err = s.collection.FindOne(ctx, bson.M{"alias": alias}).Decode(&doc); err != nil {
		if err == mongo.ErrNoDocuments {
			err = nil
			return
		}
		return
	}
	return toDomain(doc), true, nil
}

func (s *Storage) GetProjectByID(ctx context.Context, id uuid.UUID) (project domain.Project, found bool, err error) {

	var doc internal.ProjectDocument
	if err = s.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&doc); err != nil {
		if err == mongo.ErrNoDocuments {
			err = nil
			return
		}
		return
	}
	return toDomain(doc), true, nil
}

func (s *Storage) GetProjectByRepoURL(ctx context.Context, repoURL string) (project domain.Project, found bool, err error) {

	var doc internal.ProjectDocument
	if err = s.collection.FindOne(ctx, bson.M{"repo_url": repoURL}).Decode(&doc); err != nil {
		if err == mongo.ErrNoDocuments {
			err = nil
			return
		}
		return
	}
	return toDomain(doc), true, nil
}

func (s *Storage) CreateProject(ctx context.Context, project domain.Project) (id uuid.UUID, err error) {

	project.CreatedAt = time.Now()
	project.UpdatedAt = time.Now()

	doc := toDocument(project)
	if _, err = s.collection.InsertOne(ctx, doc); err != nil {
		if mongo.IsDuplicateKeyError(err) {
			return uuid.Nil, storage.ErrProjectAlreadyExists
		}
		return uuid.Nil, err
	}

	if err = s.bumpCatalogVersion(ctx, bumpMinor); err != nil {
		return uuid.Nil, err
	}

	return project.ID, nil
}

func (s *Storage) UpdateProject(ctx context.Context, alias string, project domain.Project) (err error) {

	project.UpdatedAt = time.Now()

	updateDoc := toUpdateDocument(project)
	var updateFields bson.M
	var bsonBytes []byte
	if bsonBytes, err = bson.Marshal(updateDoc); err != nil {
		return
	}
	if err = bson.Unmarshal(bsonBytes, &updateFields); err != nil {
		return
	}

	update := bson.M{
		"$set": updateFields,
	}
	var result *mongo.UpdateResult
	if result, err = s.collection.UpdateOne(ctx, bson.M{"alias": alias}, update); err != nil {
		return
	}
	if result.MatchedCount == 0 {
		return storage.ErrProjectNotFound
	}

	if err = s.bumpCatalogVersion(ctx, bumpPatch); err != nil {
		return
	}

	return
}

func (s *Storage) DeleteProject(ctx context.Context, alias string) (err error) {

	var result *mongo.DeleteResult
	if result, err = s.collection.DeleteOne(ctx, bson.M{"alias": alias}); err != nil {
		return
	}
	if result.DeletedCount == 0 {
		return storage.ErrProjectNotFound
	}

	if err = s.bumpCatalogVersion(ctx, bumpMajor); err != nil {
		return
	}

	return
}

func (s *Storage) ListProjects(ctx context.Context, limit int, offset int) (projects []domain.Project, total int64, err error) {

	if total, err = s.collection.CountDocuments(ctx, bson.M{}); err != nil {
		return
	}

	opts := options.Find().
		SetLimit(int64(limit)).
		SetSkip(int64(offset)).
		SetSort(bson.D{{Key: "created_at", Value: -1}})

	var cursor *mongo.Cursor
	if cursor, err = s.collection.Find(ctx, bson.M{}, opts); err != nil {
		return
	}
	defer cursor.Close(ctx)

	var docs []internal.ProjectDocument
	if err = cursor.All(ctx, &docs); err != nil {
		return
	}

	projects = make([]domain.Project, len(docs))
	for i := range docs {
		projects[i] = toDomain(docs[i])
	}

	return
}

func (s *Storage) Exists(ctx context.Context, alias string) (exists bool, err error) {

	var count int64
	if count, err = s.collection.CountDocuments(ctx, bson.M{"alias": alias}); err != nil {
		return
	}
	exists = count > 0
	return
}

func (s *Storage) GetProjectToken(ctx context.Context, alias string) (token string, err error) {

	var doc internal.ProjectDocument
	if err = s.collection.FindOne(ctx, bson.M{"alias": alias}, options.FindOne().SetProjection(bson.M{FieldEncryptedToken: 1})).Decode(&doc); err != nil {
		if err == mongo.ErrNoDocuments {
			return "", storage.ErrProjectNotFound
		}
		return
	}
	token = doc.EncryptedToken
	return
}
