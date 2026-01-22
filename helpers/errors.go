package helpers

import (
	"errors"
	"fmt"
	"log/slog"
	"strconv"
	"strings"

	"github.com/seniorGolang/tg-proxy/errs"
)

// Must проверяет ошибку и вызывает panic с логированием, если ошибка не nil
// err - ошибка для проверки
// msg - сообщение для логирования
// args - дополнительные типизированные аргументы для структурированного логирования (slog.Attr)
func Must(err error, msg string, args ...slog.Attr) {

	if err != nil {
		attrs := append([]slog.Attr{slog.Any("error", err)}, args...)
		anyArgs := make([]any, len(attrs))
		for i := range attrs {
			anyArgs[i] = attrs[i]
		}
		slog.Error(msg, anyArgs...)
		panic(err)
	}
}

// ExtractStatusCode извлекает HTTP статус-код из ошибки GitLab/GitHub API
// err - ошибка для проверки
// Возвращает статус-код и true, если код найден, иначе 0 и false
func ExtractStatusCode(err error) (statusCode int, found bool) {

	if err == nil {
		return 0, false
	}

	// Проверяем, является ли это ошибкой API через errors.Is
	if !errors.Is(err, errs.ErrGitLabAPI) && !errors.Is(err, errs.ErrGitHubAPI) {
		return 0, false
	}

	errStr := err.Error()

	// Ищем паттерн "status 404" или "status 500" и т.д.
	parts := strings.Split(errStr, "status ")
	if len(parts) < 2 {
		return 0, false
	}

	statusStr := strings.TrimSpace(parts[1])
	// Убираем возможные дополнительные символы после числа
	statusStr = strings.Fields(statusStr)[0]

	var parseErr error
	if statusCode, parseErr = strconv.Atoi(statusStr); parseErr != nil {
		return 0, false
	}

	return statusCode, true
}

// GetErrorMessage возвращает понятное сообщение об ошибке для клиента
// err - ошибка для обработки
// Возвращает сообщение об ошибке
func GetErrorMessage(err error) (message string) {

	if err == nil {
		return "Unknown error"
	}

	// Проверяем известные ошибки домена
	if errors.Is(err, errs.ErrProjectNotFound) {
		return "Project not found"
	}
	if errors.Is(err, errs.ErrVersionNotFound) {
		return "Version not found"
	}
	if errors.Is(err, errs.ErrVersionMismatch) {
		return "Version mismatch"
	}
	if errors.Is(err, errs.ErrProjectAlreadyExists) {
		return "Project already exists"
	}
	if errors.Is(err, errs.ErrManifestParseError) {
		return "Failed to parse manifest"
	}
	if errors.Is(err, errs.ErrManifestMarshalError) {
		return "Failed to marshal manifest"
	}

	// Проверяем ошибки API
	if errors.Is(err, errs.ErrGitLabAPI) || errors.Is(err, errs.ErrGitHubAPI) {
		if statusCode, found := ExtractStatusCode(err); found {
			switch statusCode {
			case 404:
				return "Resource not found in source"
			case 401, 403:
				return "Authentication failed or access denied"
			case 500, 502, 503:
				return "Source service unavailable"
			default:
				return fmt.Sprintf("Source API error: status %d", statusCode)
			}
		}
		return "Source API error"
	}

	// Возвращаем общее сообщение для неизвестных ошибок
	return "Internal server error"
}
