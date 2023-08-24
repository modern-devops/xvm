package mirrors

import (
	"fmt"
	"os"
	"strings"
)

const (
	zip = "zip"
	tar = "tar.gz"
)

type Mirror interface {
	GetURL(v string) (string, error)
	Versions() ([]string, error)
	BaseURL() string
}

func overwriteMirror(sdk string, mirror string) string {
	if m := os.Getenv(fmt.Sprintf("XVM_%s_MIRROR", strings.ToUpper(sdk))); m != "" {
		mirror = m
	}
	return strings.TrimSuffix(mirror, "/")
}
