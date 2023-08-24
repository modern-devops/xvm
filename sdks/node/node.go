package node

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/modern-devops/xvm/mirrors"
	"github.com/modern-devops/xvm/sdks"
	"github.com/modern-devops/xvm/tools"

	"github.com/rs/zerolog/log"
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
		PostInstall: func(wp string) error {
			return n.configNpm(wp, prefix)
		},
	}
}

// DetectVersion try to detect the node version
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

func (n *nvm) configNpm(wp string, prefix string) error {
	if err := os.MkdirAll(prefix, 0755); err != nil {
		return fmt.Errorf("failed to make directry: %s, %w", prefix, err)
	}
	nodePath := filepath.Join(wp, nodeCommandFile())
	npmJs := filepath.Join(wp, "node_modules", "npm", "bin", "npm-cli.js")
	out, err := exec.Command(nodePath, npmJs, "config", "set", "prefix", prefix).Output()
	if err != nil {
		log.Error().Msg(string(out))
		return fmt.Errorf("failed to set prefix, %w", err)
	}
	npmCache := filepath.Join(n.home, node, "npm-cache")
	if err := os.MkdirAll(npmCache, 0755); err != nil {
		return fmt.Errorf("failed to make directry: %s, %w", npmCache, err)
	}
	out, err = exec.Command(nodePath, npmJs, "config", "set", "cache", npmCache).Output()
	if err != nil {
		log.Error().Msg(string(out))
		return fmt.Errorf("failed to set cache, %w", err)
	}
	return nil
}
