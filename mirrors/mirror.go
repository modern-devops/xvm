package mirrors

import (
	"fmt"
	"os"
	"runtime"
	"strings"

	"github.com/modern-devops/xvm/tools"
)

const (
	zip = "zip"
	tar = "tar.gz"
)

type VersionDesc struct {
	URL      string `json:"url"`
	Filename string `json:"filename"`
	Version  string `json:"version"`
	Sha256   string `json:"sha256"`
}

type Mirror interface {
	Versions() ([]*VersionDesc, error)
	BaseURL() string
}

func overwriteMirror(sdk, mirror string) string {
	return strings.TrimSuffix(overwriteConfig(sdk, "mirror", mirror), "/")
}

func overwriteAPI(sdk, api string) string {
	return overwriteConfig(sdk, "api", api)
}

func overwriteConfig(sdk, key, value string) string {
	if m := os.Getenv(strings.ToUpper(fmt.Sprintf("XVM_%s_%s", sdk, key))); m != "" {
		return m
	}
	return value
}

func defaultExtension() string {
	if tools.IsWindows() {
		return zip
	}
	return tar
}

func isArmMac() bool {
	return runtime.GOARCH == "arm64" && runtime.GOOS == "darwin"
}
