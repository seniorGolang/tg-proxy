package model

type Manifest struct {
	Version   string        `yaml:"version" json:"version"`
	Manifests []ManifestRef `yaml:"manifests,omitempty" json:"manifests,omitempty"`
	Packages  []Package     `yaml:"packages,omitempty" json:"packages,omitempty"`
}

type ManifestRef struct {
	URL string `yaml:"url" json:"url"`
}

type Package struct {
	Name         string             `yaml:"name" json:"name"`
	Descr        string             `yaml:"descr,omitempty" json:"descr,omitempty"`
	Downloads    []PlatformDownload `yaml:"downloads" json:"downloads"`
	Files        []FileInstallation `yaml:"files" json:"files"`
	Scripts      *Scripts           `yaml:"scripts,omitempty" json:"scripts,omitempty"`
	Dependencies []string           `yaml:"dependencies,omitempty" json:"dependencies,omitempty"`
}

type PlatformDownload struct {
	OS   string `yaml:"os,omitempty" json:"os,omitempty"`
	Arch string `yaml:"arch,omitempty" json:"arch,omitempty"`
	URL  string `yaml:"url" json:"url"`
}

type FileInstallation struct {
	File        string `yaml:"file,omitempty" json:"file,omitempty"`
	Source      string `yaml:"source,omitempty" json:"source,omitempty"`
	Destination string `yaml:"destination" json:"destination"`
	Checksum    string `yaml:"checksum,omitempty" json:"checksum,omitempty"`
}

type Scripts struct {
	PreInstall    *ScriptAction `yaml:"pre_install,omitempty" json:"pre_install,omitempty"`
	PostInstall   *ScriptAction `yaml:"post_install,omitempty" json:"post_install,omitempty"`
	PreUninstall  *ScriptAction `yaml:"pre_uninstall,omitempty" json:"pre_uninstall,omitempty"`
	PostUninstall *ScriptAction `yaml:"post_uninstall,omitempty" json:"post_uninstall,omitempty"`
}

type ScriptAction struct {
	Script string `yaml:"script,omitempty" json:"script,omitempty"`
	Source string `yaml:"source,omitempty" json:"source,omitempty"`
	Exec   string `yaml:"exec" json:"exec"`
}

// PackageWithSource — пакет с указанием источника (агрегированный манифест).
type PackageWithSource struct {
	Package
	SourceAlias   string `json:"source_alias,omitempty"`
	SourceVersion string `json:"source_version,omitempty"`
}

// ManifestAggregatedResponse — ответ агрегированного манифеста.
type ManifestAggregatedResponse struct {
	Version  string              `json:"version"`
	Packages []PackageWithSource `json:"packages"`
}
