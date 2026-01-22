package tgproxy

// Proxy предоставляет HTTP API для прокси
type Proxy struct {
	engine     engine
	baseURL    string
	publicAuth AuthProvider
	adminAuth  AuthProvider
}

// ProxyOption опция для настройки Proxy
type ProxyOption func(*Proxy)

// PublicAuth устанавливает провайдер авторизации для публичных роутов
// Публичные роуты: получение манифестов, файлов, списка версий
func PublicAuth(auth AuthProvider) (opt ProxyOption) {
	return func(p *Proxy) {
		p.publicAuth = auth
	}
}

// AdminAuth устанавливает провайдер авторизации для админских роутов
// Админские роуты: создание, обновление, удаление проектов, список проектов
func AdminAuth(auth AuthProvider) (opt ProxyOption) {
	return func(p *Proxy) {
		p.adminAuth = auth
	}
}

// New создает новый Proxy
// engine - Core Engine для обработки запросов
// baseURL - базовый URL прокси (используется для генерации проксированных URL в манифестах)
// opts - опции для настройки Proxy (например, PublicAuth, AdminAuth для авторизации)
func New(engine engine, baseURL string, opts ...ProxyOption) (proxy *Proxy) {

	p := &Proxy{
		engine:  engine,
		baseURL: baseURL,
	}

	for _, opt := range opts {
		opt(p)
	}

	return p
}
