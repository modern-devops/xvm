//go:build windows

package binpath

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"golang.org/x/sys/windows"
	"golang.org/x/sys/windows/registry"
)

const (
	env  = "Environment"
	path = "PATH"
	xvm  = "Xvm"
)

// AddUserPath add paths to user's environment path
func AddUserPath(paths ...string) error {
	op, err := getValue(path)
	if err != nil {
		return fmt.Errorf(`failed to get value: %s, %w`, path, err)
	}
	paths = filterNewPaths(op, paths...)
	if err := addToXvmPath(paths); err != nil {
		return err
	}
	return addXvmToEnvPath()
}

func PathsPlaceholder(paths ...string) string {
	return strings.Join(paths, string(os.PathListSeparator))
}

func addToXvmPath(paths []string) error {
	xvmValue, err := getValue(xvm)
	if err != nil {
		return fmt.Errorf(`failed to get value: %s, %w`, xvm, err)
	}
	newPaths := filterNewPaths(xvmValue, paths...)
	if len(newPaths) == 0 {
		return nil
	}
	separator := string(os.PathListSeparator)
	if xvmValue != "" {
		newPaths = append(newPaths, strings.Split(xvmValue, separator)...)
	}
	if err := setValue(xvm, strings.Join(newPaths, separator)); err != nil {
		return fmt.Errorf(`failed to set value: %s, %w`, xvm, err)
	}
	return nil
}

func addXvmToEnvPath() error {
	op, err := getValue(path)
	if err != nil {
		return fmt.Errorf(`failed to get value: %s, %w`, path, err)
	}
	xvmVar := fmt.Sprintf("%%%s%%", xvm)
	if strings.Contains(op, xvmVar) {
		return nil
	}
	addXvmVar := fmt.Sprintf("%s%s%s", xvmVar, string(os.PathListSeparator), op)
	if err := setValue(path, addXvmVar); err != nil {
		return fmt.Errorf(`failed to set value: %s, %w`, path, err)
	}
	return nil
}

func setValue(key, value string) error {
	cmd := exec.Command("setx", key, value)
	return cmd.Run()
}

func getValue(key string) (string, error) {
	k, err := registry.OpenKey(registry.CURRENT_USER, env, registry.ALL_ACCESS)
	if err != nil {
		return "", fmt.Errorf(`failed to open key: HKEY_CURRENT_USER\%s, %w`, env, err)
	}
	value, _, err := k.GetStringValue(key)
	if err != nil && err != windows.ERROR_FILE_NOT_FOUND {
		return "", err
	}
	return value, nil
}

func filterNewPaths(op string, paths ...string) []string {
	dps := make([]string, 0, len(paths))
	for _, path := range paths {
		if strings.Contains(op, path) {
			continue
		}
		dps = append(dps, path)
	}
	return dps
}
