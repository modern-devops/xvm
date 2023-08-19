package env

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/modern-devops/xvm/mirrors"
	"github.com/modern-devops/xvm/tools"
	"github.com/modern-devops/xvm/tools/linker"

	"github.com/go-resty/resty/v2"
)

var logger = log.New(os.Stdout, "", log.LstdFlags)

type UserIsolatedInstaller struct {
	RootPath    string
	EnvRootPath string
	BinPath     string
}

type Installer interface {
	Install() (*Detail, error)
}

type Sdk interface {
	DetectVersion() (string, error)
	Info() (*SdkInfo, error)
}

type Command struct {
	Name string
	Path string
}

type SdkInfo struct {
	Name        string         `json:"name"`
	Commands    []Command      `json:"commands"`
	BinPaths    []string       `json:"bin-paths"`
	Mirror      mirrors.Mirror `json:"mirror"`
	PreRun      func() error   `json:"-"`
	PostInstall func() error   `json:"-"`
}

type Detail struct {
	InstallPath string
}

func (i *UserIsolatedInstaller) Install(sdk *SdkInfo, version string) (*Detail, error) {
	wp := filepath.Join(i.EnvRootPath, sdk.Name, version)
	df := filepath.Join(wp, ".done")
	if _, err := os.Stat(df); err == nil {
		return &Detail{
			InstallPath: wp,
		}, nil
	}
	if err := os.MkdirAll(wp, 0755); err != nil {
		return nil, fmt.Errorf("failed to make dir: [%s], Please check: %w", wp, err)
	}
	url, err := sdk.Mirror.GetURL(version)
	if err != nil {
		return nil, err
	}
	if err := detect(url); err != nil {
		return nil, err
	}
	if err := i.downloadAndExtracting(url, wp); err != nil {
		return nil, err
	}
	if sdk.PostInstall != nil {
		if err := sdk.PostInstall(); err != nil {
			return nil, err
		}
	}
	defer i.done(sdk, df)
	return &Detail{
		InstallPath: wp,
	}, nil
}

func (i *UserIsolatedInstaller) Link(d *SdkInfo) error {
	for _, cmd := range d.Commands {
		c := fmt.Sprintf("xvm exec %s", cmd.Name)
		if _, err := linker.New(cmd.Name, i.BinPath, c, linker.OverrideAlways); err != nil {
			return fmt.Errorf("unable to link command: %s, Please check: %w", cmd.Name, err)
		}
	}
	return nil
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
