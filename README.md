# tg-proxy

Библиотека на Go для проксирования пакетов из различных источников (GitLab, GitHub и других) через короткие алиасы. Библиотека трансформирует запросы и проксирует их к внешним API, возвращая модифицированные манифесты с обновленными URL.

> **Примечание:** Библиотека разработана для обеспечения работы проекта [tg](https://github.com/seniorGolang/tg) и может использоваться как самостоятельный компонент в других проектах.

## Особенности

- 🔌 **Модульная архитектура** — легко расширяемая через интерфейсы
- 🔐 **Безопасность** — шифрование токенов перед сохранением
- ⚡ **Производительность** — потоковая передача файлов и кеширование
- 🔄 **Множественные источники** — поддержка GitLab, GitHub и других через единый интерфейс
- 💾 **Гибкое хранилище** — поддержка MongoDB и SQL баз (PostgreSQL, SQLite, MySQL, SQL Server) через GORM
- 🌐 **HTTP API** — интеграция с `net/http` и `go-fiber`
- 🔑 **Авторизация** — раздельная авторизация для публичных и админских роутов

## Требования

- Go 1.25 или выше

## Установка

```bash
go get github.com/seniorGolang/tg-proxy
```

## Быстрый старт

```go
package main

import (
	"log/slog"
	"net/http"
	"os"

	"github.com/seniorGolang/tg-proxy"
	"github.com/seniorGolang/tg-proxy/cache/memory"
	"github.com/seniorGolang/tg-proxy/core"
	"github.com/seniorGolang/tg-proxy/encryption/aes"
	"github.com/seniorGolang/tg-proxy/helpers"
	"github.com/seniorGolang/tg-proxy/source/gitlab"
	"github.com/seniorGolang/tg-proxy/storage/mongo"
)

func main() {

	helpers.SetupLogger(helpers.ParseLogLevel(os.Getenv("LOG_LEVEL")))

	mongoURI := os.Getenv("MONGO_URI")
	proxyPort := os.Getenv("PROXY_PORT")
	proxyBaseURL := os.Getenv("PROXY_BASE_URL")
	encryptionKey := os.Getenv("ENCRYPTION_KEY")
	gitlabBaseURL := os.Getenv("GITLAB_BASE_URL")

	key := aes.DeriveKeyFromString(encryptionKey)
	aesEncryptor, err := aes.NewEncryptor(key)
	helpers.Must(err, "Failed to create encryptor")

	mongoStorage, err := mongo.NewRepository(mongoURI)
	helpers.Must(err, "Failed to connect to MongoDB")

	memoryCache := memory.NewCache()

	engine := core.NewEngine(
		core.Storage(mongoStorage),
		core.Encryptor(aesEncryptor),
		core.Cache(memoryCache),
	)

	gitlabSource := gitlab.NewClient(gitlabBaseURL)
	helpers.Must(engine.RegisterSource(gitlabSource), "Failed to register source")

	proxy := tgproxy.New(engine, proxyBaseURL)

	mux := http.NewServeMux()
	proxy.SetPublicRoutes(mux, "/")
	proxy.SetAdminRoutes(mux, "/api/v1")

	slog.Info("Starting HTTP server",
		slog.String("port", proxyPort),
		slog.String("baseURL", proxyBaseURL),
	)

	helpers.Must(http.ListenAndServe(":"+proxyPort, mux), "Server failed")
}
```

## Архитектура

Библиотека построена на принципах модульности и расширяемости:

- **Core Engine** — центральный компонент, координирующий работу всех остальных компонентов
- **Project Resolver** — разрешение алиасов проектов с кешированием и расшифровкой токенов
- **Manifest Transformer** — трансформация манифестов с заменой URL на проксированные:
    - Замена URL файлов загрузки (`downloads`) на проксированные URL
    - Замена URL манифестов (`manifests`) на проксированные URL
    - Замена URL источников скриптов (`scripts.source`) на проксированные URL
    - Автоматическая трансформация зависимостей (`dependencies`) с поиском зарегистрированных проектов и заменой на проксированные URL
- **Source Interface** — абстракция для работы с различными источниками пакетов
- **Storage Interface** — абстракция для работы с различными хранилищами
- **Encryptor Interface** — абстракция для шифрования данных
- **Cache Interface** — абстракция для кеширования данных

## Трансформация манифестов

Библиотека автоматически трансформирует манифесты при их получении, заменяя все URL на проксированные:

### Обработка URL файлов и манифестов

Все URL файлов загрузки (`packages[].downloads[].url`), ссылок на манифесты (`manifests[].url`) и источников скриптов (`packages[].scripts.*.source`) автоматически заменяются на проксированные URL, если они принадлежат домену источника проекта.

### Обработка зависимостей

Библиотека автоматически обрабатывает зависимости пакетов (`packages[].dependencies[]`):

1. **Парсинг зависимостей** — поддерживаются следующие форматы:
    - `package` — зависимость без версии
    - `package@version` — зависимость с версией
    - `source:package` — зависимость из другого репозитория без версии
    - `source:package@version` — зависимость из другого репозитория с версией
    - `URL:package@version` — зависимость из URL репозитория (например, `https://github.com/user/repo:package@1.0.0`)

2. **Поиск проектов** — для зависимостей с URL источника:
    - URL источника нормализуется (удаляется суффикс `.git`, если присутствует)
    - Выполняется поиск зарегистрированного проекта по нормализованному URL репозитория
    - Если проект найден, зависимость заменяется на проксированный URL: `{baseURL}/{alias}:{package}@{version}` или `{baseURL}/{alias}:{package}`

3. **Сохранение исходных зависимостей** — если проект не найден или зависимость не содержит URL источника, она остается без изменений

**Пример трансформации зависимостей:**

```yaml
# Исходный манифест
packages:
  - name: mypackage
    dependencies:
      - https://github.com/user/repo:mypackage@1.0.0
      - otherpackage@2.0.0
      - github:somepackage@1.5.0

# После трансформации (если проект с repo_url="https://github.com/user/repo" зарегистрирован с alias="user-repo")
packages:
  - name: mypackage
    dependencies:
      - http://proxy:8080/user-repo:mypackage@1.0.0  # заменено на проксированный URL
      - otherpackage@2.0.0                          # без изменений (нет URL источника)
      - github:somepackage@1.5.0                    # без изменений (не URL, а имя источника)
```

### Нормализация URL репозиториев

При создании или обновлении проекта URL репозитория автоматически нормализуется: удаляется суффикс `.git`, если он присутствует. Это обеспечивает корректный поиск проектов по URL при обработке зависимостей.

**Пример:**

- Входной URL: `https://github.com/user/repo.git`
- Сохраненный URL: `https://github.com/user/repo`

## API

### Публичные роуты

Публичные роуты для доступа к манифестам и файлам:

```
GET /{prefix}/{alias}/{version}/manifest.yml     - получение манифеста
GET /{prefix}/{alias}/{version}/*                 - получение файла
GET /{prefix}/{alias}/versions                            - получение списка версий
```

**Примеры:**

```bash
# Получение манифеста
curl -H "X-API-Key: public-key" \
  http://localhost:8080/api/v1/proxy/myproject/1.0.25/manifest.yml

# Получение файла
curl -H "X-API-Key: public-key" \
  http://localhost:8080/api/v1/proxy/myproject/1.0.25/package.tar.gz

# Получение списка версий
curl -H "X-API-Key: public-key" \
  http://localhost:8080/api/v1/proxy/myproject/versions
```

### Админские роуты

Админские роуты для управления проектами:

```
GET    /{prefix}/projects           - список проектов (с пагинацией)
POST   /{prefix}/projects           - создание проекта
GET    /{prefix}/projects/{alias}   - получение проекта
PUT    /{prefix}/projects/{alias}   - обновление проекта
DELETE /{prefix}/projects/{alias}   - удаление проекта
```

**Примеры:**

```bash
# Список проектов
curl -H "X-Admin-Key: admin-key" \
  "http://localhost:8080/api/v1/admin/projects?limit=10&offset=0"

# Создание проекта
curl -X POST -H "X-Admin-Key: admin-key" \
  -H "Content-Type: application/json" \
  -d '{
    "alias": "myproject",
    "repo_url": "https://gitlab.com/foo/bar",
    "token": "glpat-xxxxxxxxxxxxxxxxxxxx",
    "description": "My project",
    "source_name": "gitlab"
  }' \
  http://localhost:8080/api/v1/admin/projects

# Получение проекта
curl -H "X-Admin-Key: admin-key" \
  http://localhost:8080/api/v1/admin/projects/myproject

# Обновление проекта
curl -X PUT -H "X-Admin-Key: admin-key" \
  -H "Content-Type: application/json" \
  -d '{
    "repo_url": "https://gitlab.com/foo/bar-new",
    "description": "Updated description"
  }' \
  http://localhost:8080/api/v1/admin/projects/myproject

# Удаление проекта
curl -X DELETE -H "X-Admin-Key: admin-key" \
  http://localhost:8080/api/v1/admin/projects/myproject
```

## Авторизация

Библиотека поддерживает раздельную авторизацию для публичных и админских роутов. Доступны следующие типовые реализации из пакета `helpers`:

### Статический ключ (API Key)

```go
publicAuth := helpers.NewStaticKeyAuth("public-key", "X-API-Key")
adminAuth := helpers.NewStaticKeyAuth("admin-key", "X-Admin-Key")
```

### HTTP Basic Authentication

```go
publicAuth := helpers.NewBasicAuth("public-user", "public-pass")
adminAuth := helpers.NewBasicAuth("admin-user", "admin-pass")
```

### JWT с RSA подписью

```go
publicAuth, err := helpers.NewJWTAuth(publicKeyPEM)
adminAuth, err := helpers.NewJWTAuth(adminKeyPEM)
```

Вы также можете реализовать собственные провайдеры авторизации, реализовав интерфейсы `AuthProvider` или `FiberAuthProvider`.

## Интеграция с go-fiber

Библиотека также поддерживает интеграцию с `go-fiber`:

```go
import (
	"github.com/gofiber/fiber/v2"
	tgproxy "github.com/seniorGolang/tg-proxy"
)

app := fiber.New()

proxy := tgproxy.New(engine, "https://proxy.example.com",
	tgproxy.PublicAuth(publicAuth),
	tgproxy.AdminAuth(adminAuth),
)

// Регистрация роутов
proxy.SetPublicRoutesFiber(app, "/api/v1/proxy")
proxy.SetAdminRoutesFiber(app, "/api/v1/admin")

app.Listen(":8080")
```

## Расширение библиотеки

### Добавление нового источника

Реализуйте интерфейс `core.Source`:

```go
type Source interface {
	Name() (name string)
	GetManifest(ctx context.Context, project domain.Project, version string) (manifest domain.Manifest, err error)
	GetFileStream(ctx context.Context, project domain.Project, version string, filename string) (stream io.ReadCloser, err error)
	GetFileResponse(ctx context.Context, project domain.Project, version string, filename string) (resp *http.Response, err error)
	GetVersions(ctx context.Context, project domain.Project) (versions []string, err error)
}
```

Опционально можно реализовать интерфейс `core.SourceInfo` для предоставления метаданных:

```go
type SourceInfo interface {
	BaseURL() (url string)
}
```

Затем зарегистрируйте источник:

```go
customSource := NewCustomSource(...)
engine.RegisterSource(customSource)
```

### Добавление нового хранилища

Реализуйте интерфейс `core.Storage` (приватный интерфейс, используйте интерфейс из пакета `storage`):

```go
type Storage interface {
	GetProject(ctx context.Context, alias string) (project domain.Project, found bool, err error)
	GetProjectByRepoURL(ctx context.Context, repoURL string) (project domain.Project, found bool, err error)
	CreateProject(ctx context.Context, project domain.Project) (err error)
	UpdateProject(ctx context.Context, alias string, project domain.Project) (err error)
	DeleteProject(ctx context.Context, alias string) (err error)
	ListProjects(ctx context.Context, limit int, offset int) (projects []domain.Project, total int64, err error)
}
```

**Примечание:** Метод `GetProjectByRepoURL` используется для поиска проектов по URL репозитория при обработке зависимостей в манифестах.

Доступные реализации:
- `storage/mongo` — MongoDB хранилище
- `storage/gorm` — универсальное хранилище для SQL баз (PostgreSQL, SQLite, MySQL, SQL Server)

### Добавление нового шифратора

Реализуйте интерфейс `core.Encryptor` (приватный интерфейс):

```go
type Encryptor interface {
	EncryptString(plaintext string) (ciphertext string, err error)
	DecryptString(ciphertext string) (plaintext string, err error)
}
```

Доступные реализации:
- `encryption/aes` — AES шифрование

### Добавление нового кеша

Реализуйте интерфейс `core.Cache` (приватный интерфейс):

```go
type Cache interface {
	GetProject(ctx context.Context, alias string) (project domain.Project, found bool, err error)
	SetProject(ctx context.Context, alias string, project domain.Project, ttl time.Duration) (err error)
	DeleteProject(ctx context.Context, alias string) (err error)
	GetVersions(ctx context.Context, alias string) (versions []string, found bool, err error)
	SetVersions(ctx context.Context, alias string, versions []string, ttl time.Duration) (err error)
	Clear(ctx context.Context) (err error)
}
```

Доступные реализации:
- `cache/memory` — in-memory кеш
- `cache/redis` — Redis кеш

## Логирование

Библиотека использует структурированное типизированное логирование через стандартный пакет `log/slog`. Все логирование использует типизированные ключи через функции `slog.String()`, `slog.Int()`, `slog.Duration()`, `slog.Any()` и т.д.

**Настройка логгера:**

```go
import (
	"log/slog"
	"github.com/seniorGolang/tg-proxy/helpers"
)

helpers.SetupLogger(helpers.ParseLogLevel(os.Getenv("LOG_LEVEL")))
```

**Использование в коде:**

```go
import "log/slog"

slog.Info("Starting server",
	slog.Int("port", 8080),
	slog.String("env", "production"),
)

slog.Error("Failed to connect",
	slog.Any("error", err),
	slog.String("uri", dbURI),
)
```

## Структура проекта

```
tg-proxy/
├── api/                    # API документация
│   └── swagger.json        # OpenAPI 3.0 спецификация
├── cache/                  # Реализации кеша
│   ├── memory/            # In-memory кеш
│   └── redis/             # Redis кеш
├── core/                   # Основные компоненты библиотеки
│   ├── cache.go           # Интерфейс кеша
│   ├── encryptor.go       # Интерфейс шифратора
│   ├── engine.go          # Core Engine
│   ├── resolver.go        # Project Resolver
│   ├── source.go          # Интерфейс источника
│   ├── storage.go         # Интерфейс хранилища
│   └── transformer.go     # Manifest Transformer
├── encryption/             # Реализации шифрования
│   └── aes/               # AES шифрование
├── errs/                   # Определения ошибок
├── helpers/                # Вспомогательные функции
│   ├── auth.go            # Реализации авторизации
│   ├── errors.go          # Обработка ошибок
│   ├── logger.go          # Настройка логгера
│   ├── logkeys.go         # Ключи для логирования
│   ├── mask.go            # Маскирование токенов
│   ├── url.go             # Утилиты для URL
│   └── validator.go       # Валидация данных
├── model/                  # Модели данных
│   ├── converter.go       # Конвертеры моделей
│   ├── domain/            # Доменные модели
│   ├── dto/               # DTO модели для API
│   └── manifest.go        # Модель манифеста
├── source/                 # Реализации источников
│   ├── gitlab/            # GitLab Source Implementation
│   └── github/            # GitHub Source Implementation
├── storage/                # Реализации хранилища
│   ├── gorm/              # GORM хранилище (PostgreSQL, SQLite, MySQL, SQL Server)
│   └── mongo/             # MongoDB хранилище
├── auth.go                # Интерфейсы авторизации
├── engine.go              # Интерфейс Engine для Proxy
├── handlers.go            # HTTP обработчики
├── middleware.go          # Middleware для авторизации
├── proxy.go               # Основной Proxy компонент
└── routes.go              # Регистрация роутов
```

## Лицензия

См. файл [LICENSE](LICENSE) для подробностей.

## Вклад

Вклад в проект приветствуется! Пожалуйста, создавайте issue или pull request для предложений по улучшению.

## Документация

Основная документация находится в файле `README.md`. API документация доступна в формате OpenAPI 3.0 в файле `api/swagger.json`.
