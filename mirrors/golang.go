package mirrors

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"runtime"
)

type goMirror struct {
	GoBaseMirror string `json:"Base"`
}

const (
	zip = "zip"
	tar = "tar.gz"
)

func Go() Mirror {
	mirror := "https://go.dev/dl/"
	if m := os.Getenv("XVM_GO_MIRROR"); m != "" {
		mirror = m
	}
	return &goMirror{
		GoBaseMirror: mirror,
	}
}

func (g *goMirror) GetURL(v string) (string, error) {
	pkg := tar
	if runtime.GOOS == "windows" {
		pkg = zip
	}
	os := fmt.Sprintf("%s-%s", runtime.GOOS, runtime.GOARCH)
	return g.getFullGoURL(fmt.Sprintf("go%s.%s.%s", v, os, pkg)), nil
}

func (g *goMirror) GetLatestURL() (string, error) {
	latest, err := g.getGoVersions()
	if err != nil {
		return "", err
	}
	return g.GetURL(latest[0])
}

func (g *goMirror) getGoVersions() ([]string, error) {
	versionListURL := g.GoBaseMirror + "?mode=json&include=all"
	rsp, err := http.Get(versionListURL)
	if err != nil {
		return nil, fmt.Errorf("failed to request: %s, %w", versionListURL, err)
	}
	defer rsp.Body.Close()
	body, err := io.ReadAll(rsp.Body)
	if err != nil {
		return nil, err
	}
	var items []map[string]interface{}
	if err := json.Unmarshal(body, &items); err != nil {
		return nil, err
	}
	versions := make([]string, 0, len(items))
	for _, item := range items {
		for k, v := range item {
			if k == "version" {
				versions = append(versions, v.(string)[2:])
			}
		}
	}
	return versions, nil
}

// getFullGoURL 获取go完整的下载路径
func (g *goMirror) getFullGoURL(path string) string {
	return g.GoBaseMirror + path
}
