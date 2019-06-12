package types

import (
	"io"
	"text/template"
)

var versionTemplate = ` Version:      {{.Version}}
 Git commit:   {{.GitCommit}}
 Go version:   {{.GoVersion}}
 Built:        {{.BuildTime}}
 OS/Arch:      {{.Os}}/{{.Arch}}
`

// Version is exported
type Version struct {
	Version   string `json:"version"`
	GoVersion string `json:"go_version"`
	GitCommit string `json:"git_commit"`
	BuildTime string `json:"build_time"`
	Os        string `json:"os"`
	Arch      string `json:"arch"`
}

// WriteTo is exported
func (v Version) WriteTo(w io.Writer) (int64, error) {
	tmpl, _ := template.New("version").Parse(versionTemplate)
	return -1, tmpl.Execute(w, v) // just make pass govet
}
