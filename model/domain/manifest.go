package domain

// Manifest описывает конкретную версию релиза с пакетами и/или ссылками на другие манифесты.
type Manifest struct {
	Version   string
	Packages  []Package
	Manifests []ManifestRef
}

// ManifestRef представляет ссылку на другой манифест.
type ManifestRef struct {
	URL string
}

// Package описывает пакет из релиза.
type Package struct {
	Name         string
	Descr        string
	Downloads    []PlatformDownload
	Files        []FileInstallation
	Scripts      *Scripts
	Dependencies []string
}

// PlatformDownload содержит информацию о загрузке.
type PlatformDownload struct {
	OS   string
	Arch string
	URL  string
}

// FileInstallation описывает установку файла.
type FileInstallation struct {
	File        string
	Source      string
	Destination string
	Checksum    string
}

// Scripts содержит скрипты для выполнения.
type Scripts struct {
	PreInstall    *ScriptAction
	PostInstall   *ScriptAction
	PreUninstall  *ScriptAction
	PostUninstall *ScriptAction
}

// ScriptAction описывает действие скрипта.
type ScriptAction struct {
	Script string
	Source string
	Exec   string
}
