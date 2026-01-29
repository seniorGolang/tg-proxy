package model

import "github.com/seniorGolang/tg-proxy/model/domain"

func (m *Manifest) ToDomain() (manifest domain.Manifest) {
	manifest.Version = m.Version
	manifest.Manifests = make([]domain.ManifestRef, len(m.Manifests))
	for i := range m.Manifests {
		manifest.Manifests[i] = domain.ManifestRef{
			URL: m.Manifests[i].URL,
		}
	}
	manifest.Packages = make([]domain.Package, len(m.Packages))
	for i := range m.Packages {
		manifest.Packages[i] = m.Packages[i].ToDomain()
	}
	return
}

func (m *Manifest) FromDomain(manifest domain.Manifest) {
	m.Version = manifest.Version
	m.Manifests = make([]ManifestRef, len(manifest.Manifests))
	for i := range manifest.Manifests {
		m.Manifests[i] = ManifestRef{
			URL: manifest.Manifests[i].URL,
		}
	}
	m.Packages = make([]Package, len(manifest.Packages))
	for i := range manifest.Packages {
		m.Packages[i].FromDomain(manifest.Packages[i])
	}
}

func (p *Package) ToDomain() (pkg domain.Package) {
	pkg.Name = p.Name
	pkg.Descr = p.Descr
	pkg.Downloads = make([]domain.PlatformDownload, len(p.Downloads))
	for i := range p.Downloads {
		pkg.Downloads[i] = domain.PlatformDownload{
			OS:   p.Downloads[i].OS,
			Arch: p.Downloads[i].Arch,
			URL:  p.Downloads[i].URL,
		}
	}
	pkg.Files = make([]domain.FileInstallation, len(p.Files))
	for i := range p.Files {
		pkg.Files[i] = domain.FileInstallation{
			File:        p.Files[i].File,
			Source:      p.Files[i].Source,
			Destination: p.Files[i].Destination,
			Checksum:    p.Files[i].Checksum,
		}
	}
	if p.Scripts != nil {
		pkg.Scripts = p.Scripts.ToDomain()
	}
	pkg.Dependencies = make([]string, len(p.Dependencies))
	copy(pkg.Dependencies, p.Dependencies)
	return
}

func (p *Package) FromDomain(pkg domain.Package) {
	p.Name = pkg.Name
	p.Descr = pkg.Descr
	p.Downloads = make([]PlatformDownload, len(pkg.Downloads))
	for i := range pkg.Downloads {
		p.Downloads[i] = PlatformDownload{
			OS:   pkg.Downloads[i].OS,
			Arch: pkg.Downloads[i].Arch,
			URL:  pkg.Downloads[i].URL,
		}
	}
	p.Files = make([]FileInstallation, len(pkg.Files))
	for i := range pkg.Files {
		p.Files[i] = FileInstallation{
			File:        pkg.Files[i].File,
			Source:      pkg.Files[i].Source,
			Destination: pkg.Files[i].Destination,
			Checksum:    pkg.Files[i].Checksum,
		}
	}
	if pkg.Scripts != nil {
		p.Scripts = &Scripts{}
		p.Scripts.FromDomain(pkg.Scripts)
	}
	p.Dependencies = make([]string, len(pkg.Dependencies))
	copy(p.Dependencies, pkg.Dependencies)
}

func (s *Scripts) ToDomain() (scripts *domain.Scripts) {
	scripts = &domain.Scripts{}
	if s.PreInstall != nil {
		scripts.PreInstall = &domain.ScriptAction{
			Script: s.PreInstall.Script,
			Source: s.PreInstall.Source,
			Exec:   s.PreInstall.Exec,
		}
	}
	if s.PostInstall != nil {
		scripts.PostInstall = &domain.ScriptAction{
			Script: s.PostInstall.Script,
			Source: s.PostInstall.Source,
			Exec:   s.PostInstall.Exec,
		}
	}
	if s.PreUninstall != nil {
		scripts.PreUninstall = &domain.ScriptAction{
			Script: s.PreUninstall.Script,
			Source: s.PreUninstall.Source,
			Exec:   s.PreUninstall.Exec,
		}
	}
	if s.PostUninstall != nil {
		scripts.PostUninstall = &domain.ScriptAction{
			Script: s.PostUninstall.Script,
			Source: s.PostUninstall.Source,
			Exec:   s.PostUninstall.Exec,
		}
	}
	return
}

func (s *Scripts) FromDomain(scripts *domain.Scripts) {
	if scripts == nil {
		return
	}
	if scripts.PreInstall != nil {
		s.PreInstall = &ScriptAction{
			Script: scripts.PreInstall.Script,
			Source: scripts.PreInstall.Source,
			Exec:   scripts.PreInstall.Exec,
		}
	}
	if scripts.PostInstall != nil {
		s.PostInstall = &ScriptAction{
			Script: scripts.PostInstall.Script,
			Source: scripts.PostInstall.Source,
			Exec:   scripts.PostInstall.Exec,
		}
	}
	if scripts.PreUninstall != nil {
		s.PreUninstall = &ScriptAction{
			Script: scripts.PreUninstall.Script,
			Source: scripts.PreUninstall.Source,
			Exec:   scripts.PreUninstall.Exec,
		}
	}
	if scripts.PostUninstall != nil {
		s.PostUninstall = &ScriptAction{
			Script: scripts.PostUninstall.Script,
			Source: scripts.PostUninstall.Source,
			Exec:   scripts.PostUninstall.Exec,
		}
	}
}
