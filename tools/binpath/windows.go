//go:build windows
// +build windows

package binpath

import (
	"fmt"
	"os"
	"strings"

	"golang.org/x/sys/windows/registry"
)

const (
	env  = "Environment"
	path = "PATH"
)

// AddUserPath add paths to user's environment path
func AddUserPath(prior bool, paths ...string) error {
	k, err := registry.OpenKey(registry.CURRENT_USER, env, registry.ALL_ACCESS)
	if err != nil {
		return fmt.Errorf(`failed to open key: HKEY_CURRENT_USER\%s, %w`, env, err)
	}
	op, _, err := k.GetStringValue(path)
	if err != nil {
		return fmt.Errorf(`failed to get value: %s, %w`, path, err)
	}
	if err := k.SetStringValue(path, addPaths(op, paths...)); err != nil {
		return fmt.Errorf(`failed to set value: %s, %w`, path, err)
	}
	return nil
}

func addPaths(op string, paths ...string) string {
	dps := make([]string, 0, len(paths))
	for _, path := range paths {
		if strings.Contains(op, path) {
			continue
		}
		dps = append(dps, path)
	}
	separator := string(os.PathListSeparator)
	ops := strings.Split(op, separator)
	dps = append(dps, ops...)
	return strings.Join(dps, separator)
}
