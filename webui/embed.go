package webui

import (
	"embed"
	"encoding/base64"
)

//go:embed templates/*.gohtml static/*.css static/*.js static/favicon.svg static/logo.svg
var FS embed.FS

//go:embed static/favicon.svg
var faviconSVG []byte

var faviconDataURI string

func init() {

	if len(faviconSVG) > 0 {
		faviconDataURI = base64.StdEncoding.EncodeToString(faviconSVG)
	}
}
