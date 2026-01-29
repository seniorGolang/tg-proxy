package gitlab

import (
	"net/http"
	"net/url"
	"strings"
)

const (
	sourceName         = "gitlab"
	privateTokenHeader = "PRIVATE-TOKEN" // GitLab API требует именно "PRIVATE-TOKEN" (все заглавные)
	genericReleasePath = "/packages/generic/release/"
)

type Source struct {
	baseURL string
	token   string
	http    *http.Client
}

func (s *Source) Info() (name, url string) {
	return sourceName, s.baseURL
}

func NewClient(baseURL string, opts ...ClientOption) (src *Source) {

	s := &Source{
		baseURL: baseURL,
		http:    &http.Client{},
	}

	for _, opt := range opts {
		opt(s)
	}

	return s
}

func (s *Source) ParseFileURL(fileURL string) (version string, filename string, ok bool) {

	parsed, err := url.Parse(fileURL)
	if err != nil || parsed.Host == "" {
		return
	}

	baseParsed, err := url.Parse(s.baseURL)
	if err != nil || baseParsed.Host == "" {
		return
	}

	if parsed.Scheme != baseParsed.Scheme || parsed.Host != baseParsed.Host {
		return
	}

	idx := strings.Index(parsed.Path, genericReleasePath)
	if idx == -1 {
		return
	}

	after := strings.Trim(parsed.Path[idx+len(genericReleasePath):], "/")
	if after == "" {
		return
	}

	parts := strings.Split(after, "/")
	if len(parts) != 2 {
		return
	}

	version = parts[0]
	filename = parts[1]
	if version == "" || filename == "" {
		return
	}

	return version, filename, true
}
