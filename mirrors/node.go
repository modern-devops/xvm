package mirrors

import (
	"encoding/json"
	"fmt"
	"golang.org/x/mod/semver"
	"io"
	"net/http"
	"runtime"

	"github.com/modern-devops/xvm/tools"
)

type nodeMirror struct {
	NodeBaseMirror string `json:"Base"`
}

func Node() Mirror {
	return &nodeMirror{
		NodeBaseMirror: overwriteMirror("node", "https://nodejs.org/dist"),
	}
}

func (n *nodeMirror) GetURL(v string) (string, error) {
	arch := n.arch(v)
	if arch == "" {
		return "", fmt.Errorf("unsupported arch: %s, "+
			"You can access %s to confirm if your arch is supported", runtime.GOARCH, n.NodeBaseMirror)
	}
	return fmt.Sprintf("%s/%s/node-%s-%s-%s.%s", n.NodeBaseMirror, v, v, n.os(), n.arch(v), n.ext()), nil
}

func (n *nodeMirror) Versions() ([]string, error) {
	versionListURL := n.NodeBaseMirror + "/index.json"
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
				versions = append(versions, v.(string))
			}
		}
	}
	return versions, nil
}

func (n *nodeMirror) BaseURL() string {
	return n.NodeBaseMirror
}

// getFullGoURL 获取go完整的下载路径
func (n *nodeMirror) getFullNodeURL(path string) string {
	return n.NodeBaseMirror + path
}

func (n *nodeMirror) ext() string {
	if tools.IsWindows() {
		return zip
	}
	return tar
}

func (n *nodeMirror) os() string {
	if tools.IsWindows() {
		return "win"
	}
	return runtime.GOOS
}

func (n *nodeMirror) arch(ver string) string {
	// Note: Supports for darwin.arm64 after v.16.0
	if runtime.GOARCH == "arm64" && runtime.GOOS == "darwin" &&
		semver.Compare(ver, "v1.16.0") < 0 {
		return "amd64"
	}
	archMapping := map[string]string{
		"amd64":   "x64",
		"386":     "x86",
		"arm64":   "arm64",
		"ppc64le": "ppc64le",
		"s390x":   "s390x",
	}
	return archMapping[runtime.GOARCH]
}
