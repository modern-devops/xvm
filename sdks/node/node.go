package node

import (
	"fmt"
	"github.com/modern-devops/xvm/mirrors"
	"github.com/modern-devops/xvm/sdks"
	"github.com/modern-devops/xvm/tools"
	"github.com/rs/zerolog/log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

const (
	node = "node"
	npm  = "npm"
)

type nvm struct {
	home string
}

func Nvm(home string) *nvm {
	return &nvm{home: home}
}

func (n *nvm) Info() *sdks.SdkInfo {
	npmPackages := filepath.Join(n.home, node, "npm-packages")
	ncf := nodeCommandFile()
	return &sdks.SdkInfo{
		Name: node,
		Tools: []sdks.Tool{
			{
				Name: node,
				Path: ncf,
			},
			{
				Name: npm,
				Path: npmCommandFile(),
			},
		},
		BinPaths: []string{npmPackagesBinPath(npmPackages)},
		Mirror:   mirrors.Node(),
		PostInstall: func(wp string) error {
			if err := os.MkdirAll(npmPackages, 0755); err != nil {
				return fmt.Errorf("failed to make directry: %s, %w", npmPackages, err)
			}
			nodePath := filepath.Join(wp, ncf)
			npmJs := filepath.Join(wp, "node_modules", "npm", "bin", "npm-cli.js")
			out, err := exec.Command(nodePath, npmJs, "config", "set", "prefix", npmPackages).Output()
			if err != nil {
				log.Error().Msg(string(out))
				return fmt.Errorf("failed to set prefix, %w", err)
			}
			npmCache := filepath.Join(n.home, node, "npm-cache")
			if err := os.MkdirAll(npmPackages, 0755); err != nil {
				return fmt.Errorf("failed to make directry: %s, %w", npmCache, err)
			}
			out, err = exec.Command(nodePath, npmJs, "config", "set", "cache", npmCache).Output()
			if err != nil {
				log.Error().Msg(string(out))
				return fmt.Errorf("failed to set cache, %w", err)
			}
			return nil
		},
	}
}

// DetectVersion try to detect the go version
func (n *nvm) DetectVersion() (string, error) {
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
	return filepath.Join(npmPackages, "bin")
}

func nodeCommandFile() string {
	if tools.IsWindows() {
		return tools.CommandFile(node)
	}
	return filepath.Join("bin", tools.CommandFile(node))
}

func npmCommandFile() string {
	if tools.IsWindows() {
		return "npm.cmd"
	}
	return filepath.Join("bin", npm)
}
