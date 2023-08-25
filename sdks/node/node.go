package node

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/modern-devops/xvm/mirrors"
	"github.com/modern-devops/xvm/sdks"
	"github.com/modern-devops/xvm/tools"
)

const (
	node = "node"
	npm  = "npm"
)

const bin = "bin"

type nvm struct {
	home string
}

func Nvm(home string) *nvm {
	return &nvm{home: home}
}

func (n *nvm) Info() *sdks.SdkInfo {
	prefix := filepath.Join(n.home, node, "npm-packages")
	return &sdks.SdkInfo{
		Name: node,
		Tools: []sdks.Tool{
			{
				Name: node,
				Path: nodeCommandFile(),
			},
			{
				Name: npm,
				Path: npmCommandFile(),
			},
		},
		BinPaths: []string{npmPackagesBinPath(prefix)},
		Mirror:   mirrors.Node(),
		WithEnvs: func(wp string) []string {
			envPrefix := "PREFIX"
			if os.Getenv(envPrefix) == "" {
				return []string{envPrefix + "=" + prefix}
			}
			return nil
		},
		PostInstall: func(wp string) error {
			if tools.IsWindows() {
				return os.Symlink(filepath.Join(wp, nodeCommandFile()), filepath.Join(prefix, nodeCommandFile()))
			}
			return nil
		},
	}
}

// Version try to detect the node version
func (n *nvm) Version() (string, error) {
	vfs, err := tools.DetectVersionFiles(".nodeversion")
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
	return "", nil
}

func npmPackagesBinPath(npmPackages string) string {
	if tools.IsWindows() {
		return npmPackages
	}
	return filepath.Join(npmPackages, bin)
}

func nodeCommandFile() string {
	if tools.IsWindows() {
		return tools.CommandFile(node)
	}
	return filepath.Join(bin, tools.CommandFile(node))
}

func npmCommandFile() string {
	if tools.IsWindows() {
		return "npm.cmd"
	}
	return filepath.Join(bin, npm)
}

func builtinNpmRC(wp string) string {
	if tools.IsWindows() {
		return filepath.Join(wp, "node_modules", "npm", ".npmrc")
	}
	return filepath.Join(wp, "lib", "node_modules", "npm", ".npmrc")
}
