package github

import (
	"net/http"
	"net/url"
	"strings"
)

const (
	versionPrefix        = "v"
	sourceName           = "github"
	baseURL              = "https://github.com"
	apiBaseURL           = "https://api.github.com"
	releasesURL          = "https://github.com/%s/%s/releases/download/%s"
	gitInfoRefs          = "https://github.com/%s/%s.git/info/refs"
	gitService           = "?service=git-upload-pack"
	releasesDownloadPath = "/releases/download/"
)

type Source struct {
	baseURL string
	token   string
	http    *http.Client
}

func (s *Source) Info() (name, url string) {
	return sourceName, baseURL
}

func NewClient(opts ...ClientOption) (src *Source) {

	src = &Source{
		baseURL: baseURL,
		http:    &http.Client{},
	}

	for _, opt := range opts {
		opt(src)
	}

	return
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

	idx := strings.Index(parsed.Path, releasesDownloadPath)
	if idx == -1 {
		return
	}

	after := strings.Trim(parsed.Path[idx+len(releasesDownloadPath):], "/")
	if after == "" {
		return
	}

	parts := strings.Split(after, "/")
	if len(parts) < 2 {
		return
	}

	version = parts[0]
	filename = parts[len(parts)-1]
	if version == "" || filename == "" {
		return
	}

	return version, filename, true
}
