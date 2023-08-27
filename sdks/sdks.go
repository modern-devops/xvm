package sdks

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"

	"github.com/modern-devops/xvm/mirrors"
	"github.com/modern-devops/xvm/tools"
	"github.com/modern-devops/xvm/tools/linker"

	"github.com/rs/zerolog/log"
	"golang.org/x/mod/semver"
)

type UserIsolatedInstaller struct {
	RootPath     string
	SdkStashPath string
	BinPath      string
	DataPath     string
	ConfigPath   string
	Sdks         []Sdk
}

func NewUserIsolatedInstaller(home string, sdks []Sdk) *UserIsolatedInstaller {
	rp := filepath.Join(home, ".xvm")
	return &UserIsolatedInstaller{
		RootPath:     rp,
		SdkStashPath: filepath.Join(rp, "sdk"),
		BinPath:      filepath.Join(rp, "bin"),
		DataPath:     filepath.Join(rp, "data"),
		ConfigPath:   filepath.Join(rp, "config.ini"),
		Sdks:         sdks,
	}
}

type Sdk interface {
	Version() (string, error)
	Info() *SdkInfo
}

type SdkInfo struct {
	Name        string                   `json:"name"`
	Tools       []Tool                   `json:"tools"`
	BinPaths    []string                 `json:"binPaths"`
	Mirror      mirrors.Mirror           `json:"mirror"`
	WithEnvs    func(wp string) []string `json:"-"`
	PreRun      func(wp string) error    `json:"-"`
	PostInstall func(wp string) error    `json:"-"`
}

type SdkTool struct {
	Sdk  Sdk
	Tool Tool
	Root string
}

type Tool struct {
	Name string `json:"name"`
	Path string `json:"path"`
}

func (i *UserIsolatedInstaller) Install(name string) (*SdkTool, error) {
	st, err := i.getSdkTool(name)
	if err != nil {
		return nil, err
	}
	version, err := i.GetVersion(st.Sdk)
	if err != nil {
		return nil, err
	}
	var vd *mirrors.VersionDesc
	if version == "" {
		if vd, err = i.LatestVersion(st.Sdk); err != nil {
			return nil, err
		}
		version = strings.TrimPrefix(vd.Version, "v")
	}
	st.Root = filepath.Join(i.SdkStashPath, st.Info().Name, version)
	df := filepath.Join(st.Root, ".done")
	if _, err := os.Stat(df); err == nil {
		return st, nil
	}
	log.Info().Msgf("Installing %s@v%s ...", st.Info().Name, version)
	if err := os.RemoveAll(st.Root); err != nil {
		return nil, fmt.Errorf("Failed to remove dir: [%s], Please check: %w", st.Root, err)
	}
	if vd == nil {
		if vd, err = st.VersionDesc(version); err != nil {
			return nil, err
		}
	}
	// TODO(buthim): checksum
	if err := i.downloadAndExtracting(vd.URL, st.Root); err != nil {
		return nil, err
	}
	if pi := st.Info().PostInstall; pi != nil {
		log.Info().Msgf("Configuring ...")
		if err := pi(st.Root); err != nil {
			return nil, err
		}
	}
	return st, i.done(st.Info(), df)
}

func (i *UserIsolatedInstaller) Link(names ...string) error {
	for _, name := range names {
		if err := i.link(name); err != nil {
			return err
		}
	}
	return nil
}

func (i *UserIsolatedInstaller) GetSdk(name string) (Sdk, error) {
	names := make([]string, 0, len(i.Sdks))
	for _, sdk := range i.Sdks {
		names = append(names, sdk.Info().Name)
		if sdk.Info().Name == name {
			return sdk, nil
		}
	}
	return nil, fmt.Errorf("unknown sdk: %s, allows %s", name, strings.Join(names, ","))
}

func (i *UserIsolatedInstaller) GetVersion(sdk Sdk) (string, error) {
	if v := os.Getenv(fmt.Sprintf("XVM_%s_VERSION", strings.ToUpper(sdk.Info().Name))); v != "" {
		return strings.TrimPrefix(v, "v"), nil
	}
	v, err := sdk.Version()
	if err != nil {
		return "", err
	}
	return strings.TrimPrefix(v, "v"), nil
}

func (i *UserIsolatedInstaller) LatestVersion(sdk Sdk) (*mirrors.VersionDesc, error) {
	versions, err := sdk.Info().Mirror.Versions()
	if err != nil {
		return nil, err
	}
	if len(versions) == 0 {
		return nil, errors.New("no version found for thecurrent machine")
	}
	return slices.MaxFunc(versions, func(a, b *mirrors.VersionDesc) int {
		return semver.Compare(a.Version, b.Version)
	}), nil
}

func (s *SdkTool) VersionDesc(ver string) (*mirrors.VersionDesc, error) {
	versions, err := s.Sdk.Info().Mirror.Versions()
	if err != nil {
		return nil, err
	}
	if len(versions) == 0 {
		return nil, errors.New("no version found for thecurrent machine")
	}
	i := slices.IndexFunc(versions, func(desc *mirrors.VersionDesc) bool {
		return strings.TrimPrefix(desc.Version, "v") == ver
	})
	if i == -1 {
		return nil, fmt.Errorf("invalid version: %s", ver)
	}
	return versions[i], nil
}

func (s *SdkTool) Info() *SdkInfo {
	return s.Sdk.Info()
}

func (s *SdkTool) Run(ctx context.Context, args ...string) error {
	if preRun := s.Info().PreRun; preRun != nil {
		if err := preRun(s.Root); err != nil {
			return err
		}
	}
	tp := filepath.Join(s.Root, s.Tool.Path)
	tc := exec.CommandContext(ctx, tp, args...)
	if ie := s.Sdk.Info().WithEnvs; ie != nil {
		tc.Env = append(os.Environ(), s.Info().WithEnvs(s.Root)...)
	}
	tc.Stdin = os.Stdin
	tc.Stdout = os.Stdout
	tc.Stderr = os.Stderr
	return tc.Run()
}

func (i *UserIsolatedInstaller) link(name string) error {
	st, err := i.getSdkTool(name)
	if err != nil {
		return err
	}
	for _, tool := range st.Info().Tools {
		log.Info().Str("command", filepath.Join(i.BinPath, tool.Name)).Msg("Linking ...")
		command := fmt.Sprintf("xvm exec %s", tool.Name)
		if _, err := linker.New(tool.Name, i.BinPath, command, linker.OverrideAlways); err != nil {
			return fmt.Errorf("unable to link command: %s, Please check: %w", tool.Name, err)
		}
		log.Info().Msgf("The %s command has been linked, "+
			"add [%s] to env.PATH to enable it, ignore if you used the --add_binpath flag", tool.Name, i.BinPath)
	}
	return nil
}

func (i *UserIsolatedInstaller) getSdkTool(name string) (*SdkTool, error) {
	var names []string
	for _, sdk := range i.Sdks {
		if st := findSdkTool(sdk, name); st != nil {
			return st, nil
		}
		names = append(names, sdkToolNames(sdk)...)
	}
	return nil, fmt.Errorf("unknown sdk: %s, allows %s", name, strings.Join(names, ","))
}

func findSdkTool(sdk Sdk, name string) *SdkTool {
	for _, tool := range sdk.Info().Tools {
		if tool.Name != name {
			continue
		}
		return &SdkTool{Sdk: sdk, Tool: tool}
	}
	return nil
}

func sdkToolNames(sdk Sdk) []string {
	names := make([]string, 0, len(sdk.Info().Tools))
	for _, tool := range sdk.Info().Tools {
		names = append(names, tool.Name)
	}
	return names
}

func (i *UserIsolatedInstaller) downloadAndExtracting(url string, path string) error {
	temp, err := os.MkdirTemp("", "")
	if err != nil {
		return fmt.Errorf("Failed to make temp dir: %w", err)
	}
	defer func() {
		_ = os.RemoveAll(temp)
	}()
	log.Info().Msgf("Downloading %s ...", url)
	filename, err := tools.Download(url, temp)
	if err != nil {
		return fmt.Errorf("Failed to download: %w", err)
	}
	log.Info().Msgf("Extracting to %s ...", path)
	err = tools.Unarchive(filename, path)
	if err != nil {
		return fmt.Errorf("Failed to extracting: %w", err)
	}
	return nil
}

func (i *UserIsolatedInstaller) done(d *SdkInfo, df string) error {
	f, err := os.OpenFile(df, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0755)
	if err != nil {
		return err
	}
	data, err := json.Marshal(d)
	if err != nil {
		return err
	}
	_, err = f.Write(data)
	return err
}
