package mirrors

import (
	"runtime"
	"slices"

	"github.com/go-resty/resty/v2"
)

const node = "node"

type nodeMirror struct {
	BaseMirror string `json:"base"`
	API        string `json:"api"`
}

func Node() Mirror {
	return &nodeMirror{
		BaseMirror: overwriteMirror(node, "https://nodejs.org/dist"),
		API:        overwriteAPI(node, "https://nodejs.org/dist/index.json"),
	}
}

func (n *nodeMirror) Versions() ([]*VersionDesc, error) {
	var nvs []nodeVersion
	_, err := resty.New().R().SetResult(&nvs).Get(n.API)
	if err != nil {
		return nil, err
	}
	var versions []*VersionDesc
	for _, nv := range nvs {
		nf := nv.Files.Pick()
		if nf == nil {
			continue
		}
		filename := nf.VersionFilename(nv.Version)
		if filename == "" {
			continue
		}
		versions = append(versions, &VersionDesc{
			URL:      n.BaseMirror + "/" + nv.Version + "/" + filename,
			Filename: filename,
			Version:  nv.Version,
		})
	}
	return versions, nil
}

func (n *nodeMirror) BaseURL() string {
	return n.BaseMirror
}

type nodeFile string

type files []*nodeFile

func (f files) Pick() *nodeFile {
	hasArmMac := slices.ContainsFunc(f, func(file *nodeFile) bool {
		return file.IsMacArch("arm64")
	})
	i := slices.IndexFunc(f, func(file *nodeFile) bool {
		if runtime.GOOS != "darwin" || hasArmMac {
			return file.Match()
		}
		// uses amd64 instead of arm64
		return file.IsMacArch("amd64")
	})
	if i == -1 {
		return nil
	}
	return f[i]
}

func (f nodeFile) Match() bool {
	oa, ok := nfm[string(f)]
	if !ok {
		return false
	}
	return oa.os == runtime.GOOS && oa.arch == runtime.GOARCH
}

func (f nodeFile) IsMacArch(arch string) bool {
	oa, ok := nfm[string(f)]
	if !ok {
		return false
	}
	return oa.os == "darwin" && oa.arch == arch
}

func (f nodeFile) VersionFilename(version string) string {
	oa, ok := nfm[string(f)]
	if !ok {
		return ""
	}
	return "node-" + version + "-" + oa.filename
}

type nodeVersion struct {
	Version string `json:"version"`
	Files   files  `json:"files"`
}

var nfm = map[string]nodeOsArchDesc{
	"linux-arm64":   {os: "linux", arch: "arm64", filename: "linux-arm64." + tar},
	"linux-x64":     {os: "linux", arch: "arm64", filename: "linux-x64." + tar},
	"linux-ppc64le": {os: "linux", arch: "ppc64le", filename: "linux-ppc64le." + tar},
	"linux-s390x":   {os: "linux", arch: "s390x", filename: "linux-s390x." + tar},
	"osx-arm64-tar": {os: "darwin", arch: "arm64", filename: "darwin-arm64." + tar},
	"osx-x64-tar":   {os: "darwin", arch: "amd64", filename: "darwin-x64." + tar},
	"win-x64-zip":   {os: "windows", arch: "amd64", filename: "win-x64." + zip},
	"win-x86-zip":   {os: "windows", arch: "386", filename: "win-x86." + zip},
}

type nodeOsArchDesc struct {
	os       string
	arch     string
	filename string
}
