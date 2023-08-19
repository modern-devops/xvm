package main

import (
	"context"
	"errors"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/modern-devops/xvm/env"
	"github.com/modern-devops/xvm/env/golang"
)

func main() {
	sdk := os.Args[2]
	args := os.Args[3:]
	home, err := os.UserHomeDir()
	if err != nil {
		panic(err)
	}
	rootPath := filepath.Join(home, ".xvm")
	i := env.UserIsolatedInstaller{
		RootPath:    rootPath,
		EnvRootPath: filepath.Join(rootPath, "env"),
		BinPath:     filepath.Join(rootPath, "bin"),
	}
	sc, err := findSdkCommand(sdks(home), home, sdk)
	if err != nil {
		panic(err)
	}
	ver, err := sc.DetectVersion()
	if err != nil {
		panic(err)
	}
	detail, err := i.Install(sc.SdkInfo, ver)
	if err != nil {
		panic(err)
	}
	cf := filepath.Join(detail.InstallPath, sc.Path)
	cmd := exec.CommandContext(context.Background(), cf, args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	handleError(cmd.Run())
}

func sdks(home string) []env.Sdk {
	return []env.Sdk{
		golang.Gvm(home),
	}
}

type sdkCommand struct {
	env.Sdk

	SdkInfo *env.SdkInfo
	Path    string
}

func findSdkCommand(sdks []env.Sdk, home, name string) (*sdkCommand, error) {
	for _, sdk := range sdks {
		info, err := sdk.Info()
		if err != nil {
			return nil, err
		}
		for _, c := range info.Commands {
			if c.Name != name {
				continue
			}
			return &sdkCommand{sdk, info, c.Path}, nil
		}
	}
	return nil, nil
}

func handleError(err error) {
	if err == nil {
		return
	}
	var eErr *exec.ExitError
	// 优先继承进程退出码
	if ok := errors.As(err, &eErr); ok {
		os.Exit(eErr.ExitCode())
	} else {
		os.Exit(1)
	}
}
