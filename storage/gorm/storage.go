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

// Storage представляет универсальное хранилище на основе GORM
type Storage struct {
	db *gorm.DB
}

// NewRepository создает новое универсальное хранилище на основе GORM
// dialector - GORM dialector для базы данных (пользователь должен импортировать нужный драйвер)
// Примеры использования:
//   - SQLite: gorm.NewRepository(sqlite.Open("file:database.db"), &gorm.Config{})
//   - PostgreSQL: gorm.NewRepository(postgres.Open(dsn), &gorm.Config{})
//   - MySQL: gorm.NewRepository(mysql.Open(dsn), &gorm.Config{})
//   - SQL Server: gorm.NewRepository(sqlserver.Open(dsn), &gorm.Config{})
func NewRepository(dialector gorm.Dialector, config *gorm.Config) (stor *Storage, err error) {

	if config == nil {
		config = &gorm.Config{
			Logger:         logger.Default.LogMode(logger.Silent),
			TranslateError: true, // Преобразует специфичные ошибки БД в унифицированные типы GORM
		}
	} else if !config.TranslateError {
		// Если пользователь передал конфиг без TranslateError, включаем его для корректной обработки ошибок
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
		db: db,
	}

	if err = stor.initSchema(ctx); err != nil {
		_ = sqlDB.Close()
		return
	}

	return
}

func (s *Storage) initSchema(ctx context.Context) (err error) {

	if err = s.db.WithContext(ctx).AutoMigrate(&Project{}); err != nil {
		return fmt.Errorf("failed to migrate schema: %w", err)
	}

	return
}

func (s *Storage) GetProject(ctx context.Context, alias string) (project domain.Project, found bool, err error) {

	var p Project
	if err = s.db.WithContext(ctx).Where(generated.Project.Alias.Eq(alias)).First(&p).Error; err != nil {
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
	if err = s.db.WithContext(ctx).Where(generated.Project.RepoURL.Eq(repoURL)).First(&p).Error; err != nil {
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

func (s *Storage) CreateProject(ctx context.Context, project domain.Project) (err error) {

	project.CreatedAt = time.Now()
	project.UpdatedAt = time.Now()

	p := FromDomain(project)
	if err = s.db.WithContext(ctx).Create(&p).Error; err != nil {
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			return storage.ErrProjectAlreadyExists
		}
		return
	}

	return
}

func (s *Storage) UpdateProject(ctx context.Context, alias string, project domain.Project) (err error) {

	project.UpdatedAt = time.Now()
	project.Alias = alias

	p := FromDomain(project)
	omitFields := GetOmitFields()

	result := s.db.WithContext(ctx).
		Model(&Project{}).
		Where(generated.Project.Alias.Eq(alias)).
		Omit(omitFields...).
		Updates(p)

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return storage.ErrProjectNotFound
	}

	return
}

func (s *Storage) DeleteProject(ctx context.Context, alias string) (err error) {

	result := s.db.WithContext(ctx).Where(generated.Project.Alias.Eq(alias)).Delete(&Project{})
	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return storage.ErrProjectNotFound
	}

	return
}

func (s *Storage) ListProjects(ctx context.Context, limit int, offset int) (projects []domain.Project, total int64, err error) {

	if err = s.db.WithContext(ctx).Model(&Project{}).Count(&total).Error; err != nil {
		return
	}

	var pList []Project
	if err = s.db.WithContext(ctx).
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
	if err = s.db.WithContext(ctx).Model(&Project{}).Where(generated.Project.Alias.Eq(alias)).Count(&count).Error; err != nil {
		return
	}

	exists = count > 0
	return
}

func (s *Storage) GetProjectToken(ctx context.Context, alias string) (token string, err error) {

	var p Project
	if err = s.db.WithContext(ctx).
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
