package tools

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
)

func GitRootPath(wd string) string {
	cp := wd
	for {
		if _, err := os.Stat(filepath.Join(cp, ".git")); err == nil {
			return cp
		}
		lp := filepath.Join(cp, "..")
		if lp == cp {
			break
		}
		cp = lp
	}
	return wd
}

func DetectVersionFiles(vf string) ([]string, error) {
	wd, err := os.Getwd()
	if err != nil {
		return nil, err
	}
	var vfs []string
	vfs = append(vfs, filepath.Join(wd, vf))
	if rp := GitRootPath(wd); wd != rp && rp != "" {
		vfs = append(vfs, filepath.Join(rp, vf))
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}
	vfs = append(vfs, filepath.Join(home, vf))
	return vfs, nil
}

func IsWindows() bool {
	return runtime.GOOS == "windows"
}

func CommandFile(command string) string {
	if IsWindows() {
		return fmt.Sprintf("%s.exe", command)
	}
	return command
}
