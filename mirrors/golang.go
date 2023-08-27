package mirrors

import (
	"fmt"
	"runtime"
	"slices"
	"strings"

	"github.com/go-resty/resty/v2"
	"golang.org/x/mod/semver"
)

const golang = "go"

type goMirror struct {
	BaseMirror string `json:"base"`
	API        string `json:"api"`
}

func Go() Mirror {
	return &goMirror{
		BaseMirror: overwriteMirror(golang, "https://go.dev/dl"),
		API:        overwriteAPI(golang, "https://go.dev/dl/?mode=json&include=all"),
	}
}

func (g *goMirror) Versions() ([]*VersionDesc, error) {
	metadata, err := g.listVersionMetadata()
	if err != nil {
		return nil, err
	}
	var versions []*VersionDesc
	for _, v := range metadata {
		f := v.Files.Pick()
		if f == nil {
			continue
		}
		versions = append(versions, &VersionDesc{
			URL:      g.BaseMirror + "/" + f.Filename,
			Filename: f.Filename,
			Version:  f.Version.String(),
			Sha256:   f.Sha256,
		})
	}
	return versions, nil
}

func (g *goMirror) BaseURL() string {
	return g.BaseMirror
}

func (g *goMirror) listVersionMetadata() ([]goVersionMetadata, error) {
	var versions []goVersionMetadata
	_, err := resty.New().R().SetResult(&versions).Get(g.API)
	if err != nil {
		return nil, fmt.Errorf("Failed to request: %s, %w", g.API, err)
	}
	slices.DeleteFunc(versions, func(metadata goVersionMetadata) bool {
		return !metadata.Stable
	})
	return versions, nil
}

type goVersionFile struct {
	Filename string    `json:"filename"`
	Os       string    `json:"os"`
	Arch     string    `json:"arch"`
	Version  goVersion `json:"version"`
	Sha256   string    `json:"sha256"`
	Size     int       `json:"size"`
	Kind     string    `json:"kind"`
}

type goVersionFiles []*goVersionFile

type goVersionMetadata struct {
	Version goVersion      `json:"version"`
	Stable  bool           `json:"stable"`
	Files   goVersionFiles `json:"files"`
}

func (g goVersionFiles) Pick() *goVersionFile {
	i := slices.IndexFunc(g, func(file *goVersionFile) bool {
		return file.Match()
	})
	if i == -1 {
		return nil
	}
	return g[i]
}

// Match returns whether it matches the current machine
func (f *goVersionFile) Match() bool {
	if f.Os != runtime.GOOS {
		return false
	}
	if !strings.HasSuffix(f.Filename, defaultExtension()) {
		return false
	}
	if f.Arch == runtime.GOARCH {
		return true
	}
	// If the version less than v1.16.0, try to pick the darwin.amd64 as darwin.arm64
	if isArmMac() && f.Arch == "amd64" &&
		semver.Compare(f.Version.String(), "v1.16.0") < 0 {
		return true
	}
	return false
}

type goVersion string

func (g goVersion) String() string {
	v := strings.TrimPrefix(string(g), "go")
	// append the patch version If the version is x.x
	if strings.Count(v, ".") == 1 {
		v += ".0"
	}
	return "v" + v
}

func (g *goVersionMetadata) Match(version string) bool {
	return g.Version.String() == version
}
