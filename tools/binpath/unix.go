//go:build !windows

package binpath

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

// AddUserPath add paths to user's rc
func AddUserPath(paths ...string) error {
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	ct := getCurrentTerminal()
	if err := addUserPathToTerminal(ct, home, paths, true); err != nil {
		return err
	}
	tryAddSupportedTerminals(ct, home, paths)
	return nil
}

type terminal interface {
	Name() string
	Rc() string
	AddPathsShell(paths ...string) string
}

type terminalBash struct{}

func (t *terminalBash) Name() string {
	return "bash"
}

func (t *terminalBash) Rc() string {
	if runtime.GOOS == "darwin" {
		return ".bash_profile"
	}
	return ".bashrc"
}

func (t *terminalBash) AddPathsShell(paths ...string) string {
	return bashLikeShell(strings.Join(paths, ":"))
}

type terminalSh struct{}

func (t *terminalSh) Name() string {
	return "sh"
}

func (t *terminalSh) Rc() string {
	return (&terminalBash{}).Rc()
}

func (t *terminalSh) AddPathsShell(paths ...string) string {
	return bashLikeShell(strings.Join(paths, ":"))
}

type terminalZsh struct{}

func (t *terminalZsh) Name() string {
	return "zsh"
}

func (t *terminalZsh) Rc() string {
	return ".zshrc"
}

func (t *terminalZsh) AddPathsShell(paths ...string) string {
	return bashLikeShell(strings.Join(paths, ":"))
}

type terminalFish struct{}

func (t *terminalFish) Name() string {
	return "fish"
}

func (t *terminalFish) Rc() string {
	return ".config/fish/config.fish"
}

func (t *terminalFish) AddPathsShell(paths ...string) string {
	return fishShell(strings.Join(paths, " "))
}

type terminalZcsh struct{}

func (t *terminalZcsh) Name() string {
	return "zcsh"
}

func (t *terminalZcsh) Rc() string {
	return ".tcshrc"
}

func (t *terminalZcsh) AddPathsShell(paths ...string) string {
	return cshShell(strings.Join(paths, " "))
}

type terminalCsh struct{}

func (t *terminalCsh) Name() string {
	return "csh"
}

func (t *terminalCsh) Rc() string {
	return ".tcshrc"
}

func (t *terminalCsh) AddPathsShell(paths ...string) string {
	return cshShell(strings.Join(paths, " "))
}

func bashLikeShell(paths string) string {
	return fmt.Sprintf(`export PATH="%s:$PATH"`, paths)
}

func cshShell(paths string) string {
	return fmt.Sprintf(`set path = (%s $path)`, paths)
}

func fishShell(paths string) string {
	return fmt.Sprintf(`set -gx PATH %s $PATH`, paths)
}

func getSupportedTerminals() []terminal {
	var ts []terminal
	ts = append(ts, &terminalBash{}, &terminalSh{})
	ts = append(ts, &terminalZcsh{}, &terminalCsh{})
	ts = append(ts, &terminalZsh{})
	ts = append(ts, &terminalFish{})
	return ts
}

func addUserPathToTerminal(t terminal, home string, paths []string, force bool) error {
	rc := filepath.Join(home, t.Rc())
	if !force {
		// ignore if not exist when not force
		if _, err := os.Stat(rc); err != nil && os.IsNotExist(err) {
			return nil
		}
	}
	dps, err := filterNewPaths(rc, paths...)
	if err != nil {
		return err
	}
	f, err := os.OpenFile(rc, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0755)
	if err != nil {
		return err
	}
	defer f.Close()
	if _, err := f.Write([]byte(fmt.Sprintf("\n%s\n", t.AddPathsShell(dps...)))); err != nil {
		return err
	}
	return nil
}

func filterNewPaths(rc string, paths ...string) ([]string, error) {
	data, err := os.ReadFile(rc)
	if err != nil {
		if os.IsNotExist(err) {
			return paths, nil
		}
		return nil, err
	}
	dps := make([]string, 0, len(paths))
	for _, path := range paths {
		if bytes.Contains(data, []byte(path)) {
			continue
		}
		dps = append(dps, path)
	}
	return dps, nil
}

func getCurrentTerminal() terminal {
	ss := strings.Split(os.Getenv("SHELL"), "/")
	name := ss[len(ss)-1]
	ts := getSupportedTerminals()
	for _, t := range ts {
		if t.Name() == name {
			return t
		}
	}
	return ts[0]
}

func tryAddSupportedTerminals(current terminal, home string, paths []string) {
	// Note: added on zsh but uses on bash
	for _, t := range getSupportedTerminals() {
		if t == current {
			continue
		}
		if err := addUserPathToTerminal(t, home, paths, false); err != nil {
			continue
		}
	}
}
