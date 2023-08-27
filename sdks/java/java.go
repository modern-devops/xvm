package java

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/modern-devops/xvm/mirrors"
	"github.com/modern-devops/xvm/sdks"
	"github.com/modern-devops/xvm/tools"
)

const (
	java = "java"
)

const javaVersionFile = ".javaversion"

var javaTools = []string{java, "javac", "javadoc", "jsk", "jstack", "jar", "jlink", "jpackage"}

type jvm struct {
	home string
}

func Jvm(home string) *jvm {
	return &jvm{home: home}
}

func (j *jvm) Info() *sdks.SdkInfo {
	sts := make([]sdks.Tool, 0, len(javaTools))
	for _, tool := range javaTools {
		sts = append(sts, sdks.Tool{
			Name: tool,
			Path: filepath.Join("bin", tools.CommandFile(tool)),
		})
	}
	return &sdks.SdkInfo{
		Name:   java,
		Tools:  sts,
		Mirror: mirrors.Java(),
		WithEnvs: func(wp string) []string {
			return []string{"JAVA_HOME" + "=" + wp}
		},
	}
}

// Version try to detect the go version
func (j *jvm) Version() (string, error) {
	// 1. detect from .javaversion
	vfs, err := tools.DetectVersionFiles(javaVersionFile)
	if err != nil {
		return "", err
	}
	for _, vf := range vfs {
		if _, err := os.Stat(vf); err != nil {
			continue
		}
		data, err := os.ReadFile(vf)
		if err != nil {
			return "", err
		}
		return strings.TrimSpace(string(data)), nil
	}
	return "", nil
}
