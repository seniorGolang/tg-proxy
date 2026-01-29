package gorm

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"github.com/seniorGolang/tg-proxy/model/domain"
	"github.com/seniorGolang/tg-proxy/storage"
	"github.com/seniorGolang/tg-proxy/storage/gorm/generated"
)

type catalogBumpKind int

const (
	catalogBumpMajor catalogBumpKind = iota
	catalogBumpMinor
	catalogBumpPatch
)

type Storage struct {
	db                  *gorm.DB
	projectsTable       string
	catalogVersionTable string
}

func NewRepository(dialector gorm.Dialector, config *gorm.Config, opts ...Option) (stor *Storage, err error) {

	var o gormOptions
	for _, apply := range opts {
		apply(&o)
	}
	if o.projectsTable == "" {
		o.projectsTable = TableProjects
	}
	if o.catalogVersionTable == "" {
		o.catalogVersionTable = TableCatalogVersion
	}

	if config == nil {
		config = &gorm.Config{
			Logger:         logger.Default.LogMode(logger.Silent),
			TranslateError: true, // Преобразует специфичные ошибки БД в унифицированные типы GORM
		}
	} else if !config.TranslateError {
		config.TranslateError = true
	}

	var db *gorm.DB
	if db, err = gorm.Open(dialector, config); err != nil {
		return
	}

	var sqlDB *sql.DB
	if sqlDB, err = db.DB(); err != nil {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err = sqlDB.PingContext(ctx); err != nil {
		_ = sqlDB.Close()
		return
	}

	stor = &Storage{
		db:                  db,
		projectsTable:       o.projectsTable,
		catalogVersionTable: o.catalogVersionTable,
	}

	if err = stor.initSchema(ctx); err != nil {
		_ = sqlDB.Close()
		return
	}

	return
}

func (s *Storage) initSchema(ctx context.Context) (err error) {

	if err = s.db.WithContext(ctx).Table(s.projectsTable).AutoMigrate(&Project{}); err != nil {
		return fmt.Errorf("failed to migrate schema: %w", err)
	}
	if err = s.db.WithContext(ctx).Table(s.catalogVersionTable).AutoMigrate(&CatalogVersion{}); err != nil {
		return fmt.Errorf("failed to migrate schema: %w", err)
	}

	var v CatalogVersion
	if err = s.db.WithContext(ctx).Table(s.catalogVersionTable).Where("id = ?", CatalogVersionID).First(&v).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			v = CatalogVersion{ID: CatalogVersionID, Major: 1, Minor: 0, Patch: 0}
			if err = s.db.WithContext(ctx).Table(s.catalogVersionTable).Create(&v).Error; err != nil {
				return fmt.Errorf("failed to init catalog version: %w", err)
			}
		} else {
			return fmt.Errorf("failed to check catalog version: %w", err)
		}
	}

	return
}

func (s *Storage) GetProject(ctx context.Context, alias string) (project domain.Project, found bool, err error) {

	var p Project
	if err = s.db.WithContext(ctx).Table(s.projectsTable).Where(generated.Project.Alias.Eq(alias)).First(&p).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			err = nil
			return
		}
		return
	}

	project = p.ToDomain()
	found = true
	return
}

func (s *Storage) GetProjectByRepoURL(ctx context.Context, repoURL string) (project domain.Project, found bool, err error) {

	var p Project
	if err = s.db.WithContext(ctx).Table(s.projectsTable).Where(generated.Project.RepoURL.Eq(repoURL)).First(&p).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			err = nil
			return
		}
		return
	}

	project = p.ToDomain()
	found = true
	return
}

func (s *Storage) GetCatalogVersion(ctx context.Context) (version string, err error) {

	var v CatalogVersion
	if err = s.db.WithContext(ctx).Table(s.catalogVersionTable).Where("id = ?", CatalogVersionID).First(&v).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "1.0.0", nil
		}
		return
	}

	version = fmt.Sprintf("%d.%d.%d", v.Major, v.Minor, v.Patch)
	return
}

func (s *Storage) bumpCatalogVersion(ctx context.Context, tx *gorm.DB, kind catalogBumpKind) (err error) {

	var v CatalogVersion
	if err = tx.WithContext(ctx).Table(s.catalogVersionTable).Where("id = ?", CatalogVersionID).First(&v).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			v = CatalogVersion{ID: CatalogVersionID, Major: 1, Minor: 0, Patch: 0}
			if err = tx.WithContext(ctx).Table(s.catalogVersionTable).Create(&v).Error; err != nil {
				return
			}
		} else {
			return
		}
	}

	switch kind {
	case catalogBumpMajor:
		v.Major++
		v.Minor = 0
		v.Patch = 0
	case catalogBumpMinor:
		v.Minor++
		v.Patch = 0
	case catalogBumpPatch:
		v.Patch++
	}

	if err = tx.WithContext(ctx).Table(s.catalogVersionTable).Where("id = ?", CatalogVersionID).
		Updates(map[string]interface{}{"major": v.Major, "minor": v.Minor, "patch": v.Patch}).Error; err != nil {
		return
	}

	return
}

func (s *Storage) CreateProject(ctx context.Context, project domain.Project) (err error) {

	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) (txErr error) {
		project.CreatedAt = time.Now()
		project.UpdatedAt = time.Now()

		p := FromDomain(project)
		if txErr = tx.Table(s.projectsTable).Create(&p).Error; txErr != nil {
			if errors.Is(txErr, gorm.ErrDuplicatedKey) {
				return storage.ErrProjectAlreadyExists
			}
			return
		}

		return s.bumpCatalogVersion(ctx, tx, catalogBumpMinor)
	})
}

func (s *Storage) UpdateProject(ctx context.Context, alias string, project domain.Project) (err error) {

	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) (txErr error) {
		project.UpdatedAt = time.Now()
		project.Alias = alias

		p := FromDomain(project)
		omitFields := GetOmitFields()

		result := tx.
			Table(s.projectsTable).
			Where(generated.Project.Alias.Eq(alias)).
			Omit(omitFields...).
			Updates(p)

		if result.Error != nil {
			return result.Error
		}

		if result.RowsAffected == 0 {
			return storage.ErrProjectNotFound
		}

		return s.bumpCatalogVersion(ctx, tx, catalogBumpPatch)
	})
}

func (s *Storage) DeleteProject(ctx context.Context, alias string) (err error) {

	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) (txErr error) {
		result := tx.Table(s.projectsTable).Where(generated.Project.Alias.Eq(alias)).Delete(&Project{})
		if result.Error != nil {
			return result.Error
		}

		if result.RowsAffected == 0 {
			return storage.ErrProjectNotFound
		}

		return s.bumpCatalogVersion(ctx, tx, catalogBumpMajor)
	})
}

func (s *Storage) ListProjects(ctx context.Context, limit int, offset int) (projects []domain.Project, total int64, err error) {

	if err = s.db.WithContext(ctx).Table(s.projectsTable).Model(&Project{}).Count(&total).Error; err != nil {
		return
	}

	var pList []Project
	if err = s.db.WithContext(ctx).Table(s.projectsTable).
		Order(generated.Project.CreatedAt.Desc()).
		Limit(limit).
		Offset(offset).
		Find(&pList).Error; err != nil {
		return
	}

	projects = make([]domain.Project, len(pList))
	for i := range pList {
		projects[i] = pList[i].ToDomain()
	}

	return
}

func (s *Storage) Exists(ctx context.Context, alias string) (exists bool, err error) {

	var count int64
	if err = s.db.WithContext(ctx).Table(s.projectsTable).Where(generated.Project.Alias.Eq(alias)).Count(&count).Error; err != nil {
		return
	}

	exists = count > 0
	return
}

func (s *Storage) GetProjectToken(ctx context.Context, alias string) (token string, err error) {

	var p Project
	if err = s.db.WithContext(ctx).Table(s.projectsTable).
		Select(generated.Project.EncryptedToken.Column().Name).
		Where(generated.Project.Alias.Eq(alias)).
		First(&p).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "", storage.ErrProjectNotFound
		}
		return
	}

	token = p.EncryptedToken
	return
}

func (s *Storage) Close() (err error) {

	sqlDB, err := s.db.DB()
	if err != nil {
		return
	}

	return sqlDB.Close()
}
