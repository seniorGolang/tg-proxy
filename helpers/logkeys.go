package helpers

// Константы для типизированных ключей логирования
// Используются для структурированного логирования через slog

const (
	// Общие ключи
	LogKeyError     = "error"
	LogKeyMessage   = "message"
	LogKeyDuration  = "duration"
	LogKeyTimestamp = "timestamp"
	LogKeyRequestID = "request_id"
	LogKeyUserID    = "user_id"

	// HTTP ключи
	LogKeyMethod     = "method"
	LogKeyPath       = "path"
	LogKeyStatusCode = "status_code"
	LogKeyIP         = "ip"
	LogKeyUserAgent  = "user_agent"

	// Проект ключи
	LogKeyAlias       = "alias"
	LogKeyVersion     = "version"
	LogKeyFilename    = "filename"
	LogKeySource      = "source"
	LogKeyRepoURL     = "repo_url"
	LogKeyDescription = "description"

	// Кеш ключи
	LogKeyCacheHit  = "cache_hit"
	LogKeyCacheMiss = "cache_miss"
	LogKeyCacheKey  = "cache_key"
	LogKeyCacheTTL  = "cache_ttl"

	// Storage ключи
	LogKeyStorageType = "storage_type"
	LogKeyLimit       = "limit"
	LogKeyOffset      = "offset"
	LogKeyTotal       = "total"

	// Source ключи
	LogKeySourceURL     = "source_url"
	LogKeyVersionsCount = "versions_count"
	LogKeyRequestURL    = "request_url"

	// Encryption ключи
	LogKeyEncryptionType = "encryption_type"

	// Auth ключи
	LogKeyAuthSuccess  = "auth_success"
	LogKeyAuthProvider = "auth_provider"

	// Маскированные чувствительные данные
	LogKeyTokenMasked    = "token_masked"
	LogKeyKeyMasked      = "key_masked"
	LogKeyPasswordMasked = "password_masked"
	LogKeyEncryptedToken = "encrypted_token"

	// Действия
	LogKeyAction = "action"
)

// Значения для действий
const (
	ActionGetManifest    = "get_manifest"
	ActionGetFile        = "get_file"
	ActionGetVersions    = "get_versions"
	ActionCreateProject  = "create_project"
	ActionGetProject     = "get_project"
	ActionUpdateProject  = "update_project"
	ActionDeleteProject  = "delete_project"
	ActionListProjects   = "list_projects"
	ActionResolveProject = "resolve_project"
	ActionCacheGet       = "cache_get"
	ActionCacheSet       = "cache_set"
	ActionCacheDelete    = "cache_delete"
	ActionCacheClear     = "cache_clear"
	ActionEncrypt        = "encrypt"
	ActionDecrypt        = "decrypt"
	ActionAuthorize      = "authorize"
)
