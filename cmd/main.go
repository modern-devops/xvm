package main

import (
	"context"
	"errors"
	"fmt"
	"github.com/modern-devops/xvm/sdks/golang"
	"github.com/modern-devops/xvm/sdks/node"
	"github.com/rs/zerolog"
	"io"
	"os"
	"os/exec"
	"runtime"
	"slices"
	"strings"

	"github.com/modern-devops/xvm/sdks"
	"github.com/modern-devops/xvm/tools/binpath"

	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"gopkg.in/ini.v1"
)

const app = "xvm"

var version = "v0.0.0"

func main() {
	setFlags()
	setLogger()
	handleError(commandRoot.Execute())
}

var commandRoot = &cobra.Command{
	Use: app,
	Short: "Xvm is a tool that provides version management for multiple sdks, " +
		"allowing you to dynamically specify a version through a version description file without installation.",
	Example:       "xvm add golang",
	SilenceErrors: true,
	SilenceUsage:  true,
	Version:       fullVersion(),
}

var activateOpts = &struct {
	All        bool
	AddBinPath bool
}{}

var subCommandActivate = &cobra.Command{
	Use:           "activate [-a/--all] [--add_binpath] [sdks...]",
	Short:         "Activate the specified or all sdks",
	Example:       "Activate golang and node, and add binary path to the user's env.PATH: `xvm activate go node --add_binpath`",
	SilenceErrors: true,
	SilenceUsage:  true,
	RunE: func(cmd *cobra.Command, args []string) error {
		if _, err := exec.LookPath(app); err != nil {
			return fmt.Errorf("the %s command cannot be found, please make sure it is installed correctly", app)
		}
		installer, err := newInstaller()
		if err != nil {
			return err
		}
		sdkNames := supportedSdkNames(installer.Sdks)
		unsupportedSdkNames := filterSdkNames(sdkNames, args)
		if len(unsupportedSdkNames) > 0 {
			return fmt.Errorf("unsupported sdks: %s", strings.Join(unsupportedSdkNames, ","))
		}
		cfg := &config{path: installer.ConfigPath}
		if err := cfg.load(); err != nil {
			return err
		}
		unactivatedSdkNames := filterSdkNames(cfg.Sdks, args)
		cfg.Sdks = append(cfg.Sdks, unactivatedSdkNames...)
		var activateSdkNames []string
		if activateOpts.All {
			cfg.Sdks = sdkNames
			activateSdkNames = sdkNames
		} else {
			activateSdkNames = args
		}
		log.Info().Msgf("Start activating %s ...", activateSdkNames)
		if err := link(installer, activateSdkNames); err != nil {
			return err
		}
		if err := addBinPath(installer, cfg); err != nil {
			return err
		}
		log.Info().Msg("Succeeded to add all binary paths")
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
		// disable log when start to run
		log.Logger = log.Output(io.Discard)
		return st.Run(context.Background(), args[1:]...)
	},
}

var subCommandShow = &cobra.Command{
	Use:           "show",
	Short:         "Show xvm details",
	SilenceErrors: true,
	SilenceUsage:  true,
	RunE: func(cmd *cobra.Command, args []string) error {
		installer, err := newInstaller()
		if err != nil {
			return err
		}
		cfg := &config{path: installer.ConfigPath}
		if err := cfg.load(); err != nil {
			return err
		}

		log.Info().Msgf("Version: %v", fullVersion())
		log.Info().Msgf("Available Sdks: %v", supportedSdkNames(installer.Sdks))
		log.Info().Msgf("Activated Sdks: %v", cfg.Sdks)
		log.Info().Msgf("Workspace: %v", installer.RootPath)
		log.Info().Msgf("Sdk Root Path: %s", installer.SdkStashPath)
		log.Info().Msgf("Binary Paths: %v", installer.BinPath)

		for _, sdk := range installer.Sdks {
			log.Info().Msgf("Mirror for %s: %v", sdk.Info().Name, sdk.Info().Mirror.BaseURL())
		}

		return nil
	},
}

type config struct {
	path string
	Sdks []string `ini:"sdks"`
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

func newInstaller() (*sdks.UserIsolatedInstaller, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}
	installer := sdks.NewUserIsolatedInstaller(home, nil)
	installer.Sdks = []sdks.Sdk{golang.Gvm(home), node.Nvm(installer.DataPath)}
	return installer, nil
}

func filterSdkNames(supportedSdks []string, sdks []string) []string {
	var items []string
	for _, item := range sdks {
		if !slices.Contains(supportedSdks, item) {
			items = append(items, item)
		}
	}
	return items
}

func addBinPath(installer *sdks.UserIsolatedInstaller, cfg *config) error {
	if !activateOpts.AddBinPath {
		return nil
	}
	binPaths, err := getBinPath(installer, cfg)
	if err != nil {
		return err
	}
	log.Info().Msg("Try adding all binary paths to env.PATH")
	return binpath.AddUserPath(binPaths...)
}

func link(installer *sdks.UserIsolatedInstaller, sdkNames []string) error {
	if err := installer.Link(sdkNames...); err != nil {
		return err
	}
	return nil
}

func getBinPath(installer *sdks.UserIsolatedInstaller, cfg *config) ([]string, error) {
	binPaths := []string{installer.BinPath}
	log.Info().Strs(app, []string{installer.BinPath}).Msg("Found binpaths")
	for _, sn := range cfg.Sdks {
		sdk, err := installer.GetSdk(sn)
		if err != nil {
			continue
		}
		log.Info().Strs(sdk.Info().Name, sdk.Info().BinPaths).Msg("Collected binpaths")
		binPaths = append(binPaths, sdk.Info().BinPaths...)
	}
	return binPaths, nil
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
	log.Error().Msg(err.Error())
	var ee *exec.ExitError
	if ok := errors.As(err, &ee); ok {
		os.Exit(ee.ExitCode())
		return
	}
	os.Exit(1)
}

func setFlags() {
	commandRoot.AddCommand(subCommandActivate, subCommandExec, subCommandShow)
	subCommandActivate.Flags().BoolVar(&activateOpts.AddBinPath, "add_binpath", false, "Add xvm's binary path to the user's sdks.PATH, On a unix like system, all identified terminal rc files, such as ~/.bashrc and ~.zshrc, will be modified.")
	subCommandActivate.Flags().BoolVarP(&activateOpts.All, "all", "a", false, "Activate all supported sdks, execute `xvm list` to see detail")
}

func setLogger() {
	level := zerolog.InfoLevel
	if assertEnvTrue("DEBUG") {
		level = zerolog.DebugLevel
	}
	var out io.Writer = os.Stdout
	if assertEnvTrue("SILENT") {
		out = io.Discard
	}
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: out}).Level(level)
}

func assertEnvTrue(env string) bool {
	return strings.ToLower(os.Getenv(env)) == "true"
}

func fullVersion() string {
	return fmt.Sprintf("%s (%s.%s)", version, runtime.GOOS, runtime.GOARCH)
}
