package tools

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/go-resty/resty/v2"
)

const (
	manifestAPI      = "https://raw.githubusercontent.com/%s/%s/%s/versions-manifest.json"
	defaultReference = "main"
)

var ErrManifestNotFound = errors.New("manifest not found")

type Manifest struct {
	Version    string         `json:"version"`
	Stable     bool           `json:"stable"`
	ReleaseURL string         `json:"release_url"`
	Files      []manifestFile `json:"files"`
}

type manifestFile struct {
	Filename    string `json:"filename"`
	Arch        string `json:"arch"`
	Platform    string `json:"platform"`
	DownloadURL string `json:"download_url"`
}

func GetManifestFromRepo(owner, repo string) ([]Manifest, error) {
	return getManifest(owner, repo, defaultReference)
}

func getManifest(owner, repo string, ref string) ([]Manifest, error) {
	url := fmt.Sprintf(manifestAPI, owner, repo, ref)
	m := make([]Manifest, 0)
	rsp, err := resty.New().R().
		ForceContentType("application/json").
		SetResult(&m).
		Get(url)
	if err != nil {
		return nil, fmt.Errorf("faile to request manifest: %w", err)
	}
	if rsp.IsSuccess() {
		return m, nil
	}
	code := rsp.StatusCode()
	if code == http.StatusNotFound {
		return nil, ErrManifestNotFound
	}
	return nil, fmt.Errorf("fail to get manifest, status: %d, resp: %s", code, rsp.String())
}
