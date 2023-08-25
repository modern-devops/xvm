package mirrors

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"runtime"

	"github.com/modern-devops/xvm/tools"
)

type goMirror struct {
	GoBaseMirror string `json:"Base"`
}

func Go() Mirror {
	return &goMirror{
		GoBaseMirror: overwriteMirror("go", "https://go.dev/dl"),
	}
}

func (g *goMirror) GetURL(v string) (string, error) {
	pkg := tar
	if tools.IsWindows() {
		pkg = zip
	}
	os := fmt.Sprintf("%s-%s", runtime.GOOS, runtime.GOARCH)
	return g.getFullGoURL(fmt.Sprintf("/go%s.%s.%s", v, os, pkg)), nil
}

func (g *goMirror) Versions() ([]string, error) {
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
	for _, metadata := range items {
		if ver := findVersion(metadata); ver != "" {
			versions = append(versions, ver)
		}
	}
	return versions, nil
}

func (g *goMirror) BaseURL() string {
	return g.GoBaseMirror
}

func (g *goMirror) getFullGoURL(path string) string {
	return g.GoBaseMirror + path
}

func findVersion(metadata map[string]interface{}) string {
	for k, v := range metadata {
		if k == "version" {
			return v.(string)[2:]
		}
	}
	return ""
}
