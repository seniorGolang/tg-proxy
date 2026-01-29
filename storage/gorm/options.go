package gorm

type Option func(o *gormOptions)

type gormOptions struct {
	projectsTable       string
	catalogVersionTable string
}

func ProjectsTable(name string) (opt Option) {
	return func(o *gormOptions) {
		o.projectsTable = name
	}
}

func CatalogVersionTable(name string) (opt Option) {
	return func(o *gormOptions) {
		o.catalogVersionTable = name
	}
}
