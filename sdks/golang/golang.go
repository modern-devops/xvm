package golang

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/modern-devops/xvm/mirrors"
	"github.com/modern-devops/xvm/sdks"
	"github.com/modern-devops/xvm/tools"

	"golang.org/x/mod/modfile"
)

const (
	golang        = "go"
	bin           = "bin"
	goroot        = "GOROOT"
	gopath        = "GOPATH"
	goVersionFile = ".goversion"
	goModFile     = "go.mod"
)

type gvm struct {
	home string
}

func Gvm(home string) *gvm {
	return &gvm{home: home}
}

func (g *gvm) Info() *sdks.SdkInfo {
	goPath := filepath.Join(g.home, "go")
	return &sdks.SdkInfo{
		Name: golang,
		Tools: []sdks.Tool{
			{
				Name: golang,
				Path: filepath.Join(bin, tools.CommandFile(golang)),
			},
		},
		BinPaths: []string{filepath.Join(goPath, "bin")},
		Mirror:   mirrors.Go(),
		InjectEnvs: func(wp string) []string {
			return []string{goroot + "=" + wp, gopath + "=" + goPath}
		},
	}
}

// DetectVersion try to detect the go version
func (g *gvm) DetectVersion() (string, error) {
	// 1. detect from .goversion
	vfs, err := tools.DetectVersionFiles(goVersionFile)
	if err != nil {
		return "", err
	}
	for _, vf := range vfs {
		if _, err := os.Stat(vf); err != nil {
			continue
		}
		data, err := os.ReadFile(vf)
		if err != nil {
			return "", err
		}
		return strings.TrimSpace(string(data)), nil
	}
	// 2. detect from go.mod
	gms, err := detectGoMods()
	if err != nil {
		return "", err
	}
	for _, gm := range gms {
		if _, err := os.Stat(gm); err != nil {
			continue
		}
		v, err := getGoModuleVersion(gm)
		if err != nil {
			return "", err
		}
		return v, nil
	}
	return "", nil
}

func detectGoMods() ([]string, error) {
	wd, err := os.Getwd()
	if err != nil {
		return nil, err
	}
	var gms []string
	gms = append(gms, filepath.Join(wd, goModFile))
	if rp := tools.GitRootPath(wd); wd != rp {
		gms = append(gms, filepath.Join(rp, goModFile))
	}
	return gms, nil
}

func getGoModuleVersion(gm string) (string, error) {
	data, err := os.ReadFile(gm)
	if err != nil {
		return "", err
	}
	f, err := modfile.Parse(gm, data, nil)
	if err != nil {
		return "", fmt.Errorf("parse %s failed, %w", gm, err)
	}
	return f.Go.Version, nil
}
