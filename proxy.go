package tgproxy

type Proxy struct {
	engine     engine
	baseURL    string
	publicAuth AuthProvider
	adminAuth  AuthProvider
}

type ProxyOption func(*Proxy)

func PublicAuth(auth AuthProvider) (opt ProxyOption) {
	return func(p *Proxy) {
		p.publicAuth = auth
	}
}

func AdminAuth(auth AuthProvider) (opt ProxyOption) {
	return func(p *Proxy) {
		p.adminAuth = auth
	}
}

func New(engine engine, baseURL string, opts ...ProxyOption) (proxy *Proxy) {

	proxy = &Proxy{
		engine:  engine,
		baseURL: baseURL,
	}

	for _, opt := range opts {
		opt(proxy)
	}

	return
}
