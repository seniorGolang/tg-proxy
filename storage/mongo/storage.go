package mongo

import (
	"context"
	"errors"
	"net/url"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"

	"github.com/seniorGolang/tg-proxy/model/domain"
	"github.com/seniorGolang/tg-proxy/storage"
	"github.com/seniorGolang/tg-proxy/storage/mongo/internal"
)

type Storage struct {
	client     *mongo.Client
	database   string
	collection *mongo.Collection
}

// NewRepository создает новый MongoDB репозиторий
func NewRepository(uri string) (stor *Storage, err error) {

	var parsedURI *url.URL
	if parsedURI, err = url.Parse(uri); err != nil {
		return
	}

	database := strings.TrimPrefix(parsedURI.Path, "/")
	if database == "" {
		stor = nil
		err = errors.New("database name must be specified in URI path")
		return
	}

	opts := options.Client().ApplyURI(uri)
	var client *mongo.Client
	client, err = mongo.Connect(opts)
	if err != nil {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err = client.Ping(ctx, nil); err != nil {
		return
	}

	collection := client.Database(database).Collection(CollectionProjects)

	ctxIndex, cancelIndex := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancelIndex()

	if _, err = collection.Indexes().CreateMany(ctxIndex, GetIndexModels()); err != nil {
		return
	}

	stor = &Storage{
		client:     client,
		database:   database,
		collection: collection,
	}
	return
}

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

func (s *Storage) CreateProject(ctx context.Context, project domain.Project) (err error) {

	project.CreatedAt = time.Now()
	project.UpdatedAt = time.Now()

	doc := toDocument(project)
	if _, err = s.collection.InsertOne(ctx, doc); err != nil {
		if mongo.IsDuplicateKeyError(err) {
			return storage.ErrProjectAlreadyExists
		}
		return
	}
	return
}

func (s *Storage) UpdateProject(ctx context.Context, alias string, project domain.Project) (err error) {

	project.UpdatedAt = time.Now()

	updateDoc := toUpdateDocument(project)
	// Используем bson.Marshal для автоматического применения bson тегов из структуры
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
