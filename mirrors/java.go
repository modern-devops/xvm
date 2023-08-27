package mirrors

import (
	"fmt"
	"os"
	"runtime"
	"slices"
	"strconv"
	"strings"

	"github.com/go-resty/resty/v2"
	"github.com/rs/zerolog/log"
)

const (
	java = "java"
	zulu = "zulu"
)

func Java() Mirror {
	mirrors := []*distributionMirror{
		newZuluMirror(),
	}
	distribution := os.Getenv("XVM_JAVA_DISTRIBUTION")
	if distribution == "" {
		distribution = defaultDistribution()
	}
	dm := getDistributionMirror(mirrors, distribution)
	if dm == nil {
		log.Error().Msgf("unknown distribution: %s, uses %s instead.", distribution, defaultDistribution())
		dm = getDistributionMirror(mirrors, defaultDistribution())
	}
	return dm.Mirror
}

func defaultDistribution() string {
	return zulu
}

func getDistributionMirror(dms []*distributionMirror, name string) *distributionMirror {
	i := slices.IndexFunc(dms, func(mirror *distributionMirror) bool {
		return mirror.Name == name
	})
	if i == -1 {
		return nil
	}
	return dms[i]
}

type distributionMirror struct {
	Name   string
	Mirror Mirror
}

type zuluMirror struct {
	API  string `json:"api"`
	Base string `json:"base"`
}

func newZuluMirror() *distributionMirror {
	return &distributionMirror{
		Name: zulu,
		Mirror: zuluMirror{
			API:  overwriteConfig(java, "api", "https://api.azul.com/zulu/download/community/v1.0/bundles/"),
			Base: overwriteMirror(java, "https://cdn.azul.com/zulu/bin/"),
		},
	}
}

func (m zuluMirror) Versions() ([]*VersionDesc, error) {
	metadata, err := m.usableVersionMetadata()
	if err != nil {
		return nil, err
	}
	versions := make([]*VersionDesc, 0, len(metadata))
	for _, v := range metadata {
		versions = append(versions, &VersionDesc{
			URL:      v.Url,
			Filename: v.Name,
			Version:  v.String(),
		})
	}
	return versions, nil
}

func (m zuluMirror) BaseURL() string {
	return m.Base
}

type archOpts struct {
	Arch string
	Bit  string
}

type versionMetadata struct {
	Version []int  `json:"java_version"`
	Url     string `json:"url"`
	Name    string `json:"name"`
}

func (v versionMetadata) String() string {
	bits := make([]string, 0, len(v.Version))
	for _, bit := range v.Version {
		bits = append(bits, strconv.Itoa(bit))
	}
	return "v" + strings.Join(bits, ".")
}

func (m zuluMirror) usableVersionMetadata() ([]versionMetadata, error) {
	arch := m.arch()
	var versions []versionMetadata
	resp, err := resty.New().R().SetQueryParams(map[string]string{
		"os":             m.os(),
		"ext":            defaultExtension(),
		"bundle_type":    "jdk",
		"arch":           arch.Arch,
		"hw_bitness":     arch.Bit,
		"release_status": "ga",
		"javafx":         "false",
	}).SetResult(&versions).Get(m.API)
	if err != nil {
		fmt.Println(resp)
		return nil, fmt.Errorf("Failed to request: %s, %w", m.API, err)
	}
	return versions, nil
}

func (m zuluMirror) arch() *archOpts {
	switch runtime.GOARCH {
	case "amd64":
		return &archOpts{"x86", "64"}
	case "386":
		return &archOpts{"x86", "32"}
	case "arm64":
		return &archOpts{"arm", "64"}
	default:
		return &archOpts{runtime.GOARCH, ""}
	}
}

func (m zuluMirror) os() string {
	switch runtime.GOOS {
	case "darwin":
		return "macos"
	default:
		return runtime.GOOS
	}
}
