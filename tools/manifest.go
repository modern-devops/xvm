package tools

import (
	"errors"
	"fmt"

	"github.com/go-resty/resty/v2"
)

const (
	manifestAPI  = "https://raw.githubusercontent.com/%s/%s/%s/versions-manifest.json"
	mainRef      = "main"
	notFoundCode = 404
)

// ErrorManifestNotFound 未找到清单文件
var ErrorManifestNotFound = errors.New("manifest not found")

// Manifest 清单
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

// GetManifestFromRepo 获取资源仓库主干分支的版本清单
func GetManifestFromRepo(owner, repo string) ([]Manifest, error) {
	return getManifest(owner, repo, mainRef)
}

func getManifest(owner, repo string, ref string) ([]Manifest, error) {
	url := fmt.Sprintf(manifestAPI, owner, repo, ref)
	m := make([]Manifest, 0)
	rsp, err := resty.New().R().
		ForceContentType("application/json").
		SetResult(&m).
		Get(url)
	if err != nil {
		return nil, fmt.Errorf("获取清单列表失败，请检查网络：%s", err)
	}
	if rsp.IsSuccess() {
		return m, nil
	}
	code := rsp.StatusCode()
	if code == notFoundCode {
		return nil, ErrorManifestNotFound
	}
	return nil, fmt.Errorf("请求失败，响应码：%d, 响应内容：%s", code, rsp.String())
}
