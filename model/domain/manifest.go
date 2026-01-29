package domain

type Manifest struct {
	Version   string
	Packages  []Package
	Manifests []ManifestRef
}

type ManifestRef struct {
	URL string
}

type Package struct {
	Name         string
	Descr        string
	Downloads    []PlatformDownload
	Files        []FileInstallation
	Scripts      *Scripts
	Dependencies []string
}

type PlatformDownload struct {
	OS   string
	Arch string
	URL  string
}

type FileInstallation struct {
	File        string
	Source      string
	Destination string
	Checksum    string
}

type Scripts struct {
	PreInstall    *ScriptAction
	PostInstall   *ScriptAction
	PreUninstall  *ScriptAction
	PostUninstall *ScriptAction
}

type ScriptAction struct {
	Script string
	Source string
	Exec   string
}
