package tgproxy

import (
	"strings"

	"github.com/seniorGolang/tg-proxy/helpers"
)

type Proxy struct {
	engine       engine
	baseURL      string
	publicPrefix string
	publicAuth   AuthProvider
	adminAuth    AuthProvider
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

func (p *Proxy) BaseURL() (baseURL string) {
	return p.baseURL
}

func (p *Proxy) PublicPrefix() (prefix string) {
	return p.publicPrefix
}

func (p *Proxy) manifestSourceBaseURL() (baseURL string) {

	if p.publicPrefix == "" || p.publicPrefix == "/" {
		return p.baseURL
	}
	return helpers.BuildURL(p.baseURL, strings.TrimPrefix(p.publicPrefix, "/"))
}
