package mongo

type Option func(o *mongoOptions)

type mongoOptions struct {
	projectsCollection       string
	catalogVersionCollection string
	catalogVersionDocID      string
}

func ProjectsCollection(name string) (opt Option) {
	return func(o *mongoOptions) {
		o.projectsCollection = name
	}
}

func CatalogVersionCollection(name string) (opt Option) {
	return func(o *mongoOptions) {
		o.catalogVersionCollection = name
	}
}

func CatalogVersionDocID(id string) (opt Option) {
	return func(o *mongoOptions) {
		o.catalogVersionDocID = id
	}
}
