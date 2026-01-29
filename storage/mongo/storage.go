package mongo

import (
	"context"
	"errors"
	"net/url"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type Storage struct {
	client              *mongo.Client
	database            string
	collection          *mongo.Collection
	versionCollection   *mongo.Collection
	catalogVersionDocID string
}

func NewRepository(uri string, opts ...Option) (stor *Storage, err error) {

	var o mongoOptions
	for _, apply := range opts {
		apply(&o)
	}
	if o.projectsCollection == "" {
		o.projectsCollection = CollectionProjects
	}
	if o.catalogVersionCollection == "" {
		o.catalogVersionCollection = CollectionCatalogVersion
	}
	if o.catalogVersionDocID == "" {
		o.catalogVersionDocID = DocIDCatalogVersion
	}

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

	clientOpts := options.Client().ApplyURI(uri).SetRegistry(newBSONRegistryWithUUID())
	var client *mongo.Client
	client, err = mongo.Connect(clientOpts)
	if err != nil {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err = client.Ping(ctx, nil); err != nil {
		return
	}

	collection := client.Database(database).Collection(o.projectsCollection)
	versionCollection := client.Database(database).Collection(o.catalogVersionCollection)

	ctxIndex, cancelIndex := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancelIndex()

	if _, err = collection.Indexes().CreateMany(ctxIndex, GetIndexModels()); err != nil {
		return
	}

	stor = &Storage{
		client:              client,
		database:            database,
		collection:          collection,
		versionCollection:   versionCollection,
		catalogVersionDocID: o.catalogVersionDocID,
	}

	if err = stor.ensureCatalogVersion(ctx); err != nil {
		return
	}

	return
}
