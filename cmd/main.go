package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"

	"github.com/modern-devops/xvm/sdks"
	"github.com/modern-devops/xvm/sdks/golang"
	"github.com/modern-devops/xvm/tools/binpath"

	"github.com/spf13/cobra"
	"gopkg.in/ini.v1"
)

func main() {
	setFlags()
	handleError(commandRoot.Execute())
}

var commandRoot = &cobra.Command{
	Use: "xvm",
	Short: "Xvm is a tool that provides version management for multiple sdks, " +
		"allowing you to dynamically specify a version through a version description file without installation.",
	Example:       "xvm add golang",
	SilenceErrors: true,
	SilenceUsage:  true,
}

var activateOpts = &struct {
	All     bool
	AddPath bool
}{}

var subCommandActivate = &cobra.Command{
	Use:           "activate [-a/--all] [--add_binpath] [sdks...]",
	Short:         "Activate the specified or all sdks",
	Example:       "Activate golang and node, and add binary path to the user's sdks.PATH: `xvm activate go node --add_path`",
	SilenceErrors: true,
	SilenceUsage:  true,
	RunE: func(cmd *cobra.Command, args []string) error {
		installer, err := newInstaller()
		if err != nil {
			return err
		}
		sdkNames := supportedSdkNames(installer.Sdks)
		unsupportedSdkNames := filterItems(sdkNames, args, exclude)
		if len(unsupportedSdkNames) > 0 {
			return fmt.Errorf("unsupported sdks: %s", strings.Join(unsupportedSdkNames, ","))
		}
		cfg := &config{path: installer.ConfigPath}
		if err := cfg.load(); err != nil {
			return err
		}
		unactivatedSdkNames := filterItems(cfg.Sdks, args, exclude)
		cfg.Sdks = append(cfg.Sdks, unactivatedSdkNames...)
		var activateSdkNames []string
		if activateOpts.All {
			cfg.Sdks = sdkNames
			activateSdkNames = sdkNames
		} else {
			activateSdkNames = args
		}
		if err := installer.Link(activateSdkNames...); err != nil {
			return err
		}
		if activateOpts.AddPath {
			if err := binpath.AddUserPath("`xvm show binpaths`"); err != nil {
				return err
			}
		}
		return cfg.save()
	},
}

var subCommandExec = &cobra.Command{
	Use:                "exec <sdk> [args...]",
	Short:              "Execute the sdk with the additional arguments",
	Example:            "Executing `xvm exec go version` is equivalent to executing `go version` directly",
	DisableFlagParsing: true,
	SilenceErrors:      true,
	SilenceUsage:       true,
	Hidden:             true,
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			return cmd.Help()
		}
		installer, err := newInstaller()
		if err != nil {
			return err
		}
		st, err := installer.Install(args[0])
		if err != nil {
			return err
		}
		tp := filepath.Join(st.InstallPath, st.Tool.Path)
		tc := exec.CommandContext(context.Background(), tp, args[1:]...)
		tc.Stdin = os.Stdin
		tc.Stdout = os.Stdout
		tc.Stderr = os.Stderr
		return tc.Run()
	},
}

var subCommandShow = &cobra.Command{
	Use:           "show",
	Short:         "Activate the specified or all sdks",
	Example:       "Activate golang and node, and add binary path to the user's sdks.PATH: `xvm activate go node --add_path`",
	SilenceErrors: true,
	SilenceUsage:  true,
}

var subCommandShowBinPaths = &cobra.Command{
	Use:           "binpaths",
	Short:         "Show all binpaths managed by xvm",
	SilenceErrors: true,
	SilenceUsage:  true,
	RunE: func(cmd *cobra.Command, args []string) error {
		installer, err := newInstaller()
		if err != nil {
			return err
		}
		c := &config{
			path: installer.ConfigPath,
		}
		if err := c.load(); err != nil {
			return err
		}
		var binpaths []string
		binpaths = append(binpaths, installer.BinPath)
		for _, sn := range c.Sdks {
			sdk, err := installer.FindSdk(sn)
			if err != nil {
				continue
			}
			binpaths = append(binpaths, sdk.Info().BinPaths...)
		}
		os.Stdout.WriteString(binpath.PathsPlaceholder(binpaths...))
		return nil
	},
}

type config struct {
	path string
	Sdks []string `ini:"sdks"`
}

type sdkConfig struct {
	Activated bool `ini:activated`
}

func (c *config) load() error {
	cfg, err := ini.LooseLoad(c.path)
	if err != nil {
		return err
	}
	return cfg.MapTo(c)
}

func (c *config) save() error {
	cfg := ini.Empty()
	if err := ini.ReflectFrom(cfg, c); err != nil {
		return err
	}
	return cfg.SaveToIndent(c.path, "\t")
}

func configFile() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".xvm", "config.ini"), nil
}

func newInstaller() (*sdks.UserIsolatedInstaller, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}
	return sdks.NewUserIsolatedInstaller(home, []sdks.Sdk{golang.Gvm(home)}), nil
}

func filterItems(litems []string, ritems []string, filter filter) []string {
	var items []string
	for _, item := range ritems {
		if filter(litems, item) {
			items = append(items, item)
		}
	}
	return items
}

type filter func([]string, string) bool

func include(items []string, item string) bool {
	return slices.Contains(items, item)
}

func exclude(items []string, item string) bool {
	return !slices.Contains(items, item)
}

func supportedSdkNames(sdks []sdks.Sdk) []string {
	names := make([]string, 0, len(sdks))
	for _, sdk := range sdks {
		names = append(names, sdk.Info().Name)
	}
	return names
}

func handleError(err error) {
	if err == nil {
		return
	}
	var ee *exec.ExitError
	if ok := errors.As(err, &ee); ok {
		os.Exit(ee.ExitCode())
		return
	}
	os.Exit(1)
}

func setFlags() {
	commandRoot.AddCommand(subCommandActivate, subCommandShow, subCommandExec)
	subCommandShow.AddCommand(subCommandShowBinPaths)
	subCommandActivate.Flags().BoolVar(&activateOpts.AddPath, "add_binpath", false, "Add xvm's binary path to the user's sdks.PATH, On a unix like system, all identified terminal rc files, such as ~/.bashrc and ~.zshrc, will be modified.")
	subCommandActivate.Flags().BoolVarP(&activateOpts.All, "all", "a", false, "Activate all supported sdks, execute `xvm list` to see detail")
}
