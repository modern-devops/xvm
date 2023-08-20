package sdks

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/modern-devops/xvm/mirrors"
	"github.com/modern-devops/xvm/tools"
	"github.com/modern-devops/xvm/tools/linker"

	"github.com/go-resty/resty/v2"
)

var logger = log.New(os.Stdout, "", log.LstdFlags)

type UserIsolatedInstaller struct {
	RootPath     string
	SdkStashPath string
	BinPath      string
	ConfigPath   string
	Sdks         []Sdk
}

func NewUserIsolatedInstaller(home string, sdks []Sdk) *UserIsolatedInstaller {
	rp := filepath.Join(home, ".xvm")
	return &UserIsolatedInstaller{
		RootPath:     rp,
		SdkStashPath: filepath.Join(rp, "sdk"),
		BinPath:      filepath.Join(rp, "bin"),
		ConfigPath:   filepath.Join(rp, "config.ini"),
		Sdks:         sdks,
	}
}

type Sdk interface {
	DetectVersion() (string, error)
	Info() *SdkInfo
}

type Tool struct {
	Name string `json:"name"`
	Path string `json:"path"`
}

type SdkInfo struct {
	Name        string         `json:"name"`
	Tools       []Tool         `json:"tools"`
	BinPaths    []string       `json:"binPaths"`
	Mirror      mirrors.Mirror `json:"mirror"`
	PreRun      func() error   `json:"-"`
	PostInstall func() error   `json:"-"`
}

type SdkTool struct {
	Sdk         Sdk
	Tool        Tool
	InstallPath string
}

func (i *UserIsolatedInstaller) Install(sdkName string) (*SdkTool, error) {
	st, err := i.findSdkTool(sdkName)
	if err != nil {
		return nil, err
	}
	version, err := st.Sdk.DetectVersion()
	if err != nil {
		return nil, err
	}
	wp := filepath.Join(i.SdkStashPath, sdkName, version)
	st.InstallPath = wp
	df := filepath.Join(wp, ".done")
	if _, err := os.Stat(df); err == nil {
		return st, nil
	}
	if err := os.MkdirAll(wp, 0755); err != nil {
		return nil, fmt.Errorf("failed to make dir: [%s], Please check: %w", wp, err)
	}
	url, err := st.Sdk.Info().Mirror.GetURL(version)
	if err != nil {
		return nil, err
	}
	if err := detect(url); err != nil {
		return nil, err
	}
	if err := i.downloadAndExtracting(url, wp); err != nil {
		return nil, err
	}
	if pi := st.Sdk.Info().PostInstall; pi != nil {
		return nil, pi()
	}
	defer i.done(st.Sdk.Info(), df)
	return st, nil
}

func (i *UserIsolatedInstaller) Link(sdkNames ...string) error {
	for _, sdk := range sdkNames {
		if err := i.link(sdk); err != nil {
			return err
		}
	}
	return nil
}

func (i UserIsolatedInstaller) link(sdkName string) error {
	st, err := i.findSdkTool(sdkName)
	if err != nil {
		return err
	}
	for _, tool := range st.Sdk.Info().Tools {
		c := fmt.Sprintf("xvm exec %s", tool.Name)
		if _, err := linker.New(tool.Name, i.BinPath, c, linker.OverrideAlways); err != nil {
			return fmt.Errorf("unable to link command: %s, Please check: %w", tool.Name, err)
		}
	}
	return nil
}

func (i *UserIsolatedInstaller) FindSdk(name string) (Sdk, error) {
	names := make([]string, 0, len(i.Sdks))
	for _, sdk := range i.Sdks {
		names = append(names, sdk.Info().Name)
		if sdk.Info().Name == name {
			return sdk, nil
		}
	}
	return nil, fmt.Errorf("unknown sdk: %s, allows %s", name, strings.Join(names, ","))
}

func (i *UserIsolatedInstaller) findSdkTool(name string) (*SdkTool, error) {
	names := make([]string, 0, len(i.Sdks))
	for _, sdk := range i.Sdks {
		for _, tool := range sdk.Info().Tools {
			if tool.Name != name {
				names = append(names, tool.Name)
				continue
			}
			return &SdkTool{Sdk: sdk, Tool: tool}, nil
		}
	}
	return nil, fmt.Errorf("unknown sdk: %s, allows %s", name, strings.Join(names, ","))
}

func (i *UserIsolatedInstaller) downloadAndExtracting(url string, path string) error {
	temp, err := os.MkdirTemp("", "")
	if err != nil {
		return fmt.Errorf("failed to make temp dir: %w", err)
	}
	defer func() {
		_ = os.RemoveAll(temp)
	}()
	logger.Printf("Downloading %s ...", url)
	filename, err := tools.Download(url, temp)
	if err != nil {
		return fmt.Errorf("failed to download: %w", err)
	}
	logger.Printf("Extracting to %s ...", path)
	err = tools.Unarchive(filename, path)
	if err != nil {
		return fmt.Errorf("failed to extracting: %w", err)
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

func detect(url string) error {
	rsp, err := resty.New().R().Head(url)
	if err != nil {
		return fmt.Errorf("failed to probe this url: [%s], Please check: %w", url, err)
	}
	if rsp != nil && rsp.StatusCode() == http.StatusNotFound {
		return fmt.Errorf("this url was not found: %s", url)
	}
	return nil
}
