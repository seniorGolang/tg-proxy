package internal

// Release представляет релиз GitHub для парсинга JSON ответа
type Release struct {
	TagName string `json:"tag_name"`
}
