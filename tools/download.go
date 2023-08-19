package tools

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-resty/resty/v2"
	"github.com/google/uuid"
)

const (
	contentDisposition = "Content-Disposition"
)

// Download 下载资源
// toPath 如果为文件夹，则下载到该文件夹
// 并尽可能地用资源服务器提供的文件命名（来自于response -> header -> Content-Disposition）
// 如果资源服务器没有提供，则文件会被命名为一串uuid
// 如果toPath为文件完整路径，则直接以toPath作为下载后的文件
func Download(url string, toPath string) (string, error) {
	filename := findFilename(url, toPath)
	rsp, err := downloadToPath(url, filename)
	if err != nil {
		return "", err
	}
	if rsp.IsError() {
		return "", fmt.Errorf("下载错误：%s", rsp.String())
	}
	// 文件夹的话，尝试用contentDisposition指明的文件名
	return tryRename(rsp, toPath, filename)
}

// 尝试获取文件名
// 如果toPath为指定文件，则使用toPath
// 否则尝试查找url
func findFilename(url string, toPath string) string {
	if fi, err := os.Stat(toPath); err != nil || !fi.IsDir() {
		return toPath
	}
	lastSlash := strings.LastIndex(url, "/")
	if lastSlash == -1 {
		return genTempFile()
	}
	filename := url[lastSlash+1:]
	firstQuestion := strings.Index(filename, "?")
	if firstQuestion != -1 {
		filename = filename[0:firstQuestion]
	}
	return filepath.Join(toPath, filename)
}

func downloadToPath(url, filename string) (*resty.Response, error) {
	return resty.New().R().SetOutput(filename).Get(url)
}

func tryRename(rsp *resty.Response, toPath, oFilename string) (string, error) {
	if fi, err := os.Stat(toPath); err != nil || !fi.IsDir() {
		return oFilename, nil
	}
	contentFilename := readFileName(rsp)
	if contentFilename == "" {
		return oFilename, nil
	}
	if strings.HasSuffix(oFilename, contentFilename) {
		return oFilename, nil
	}
	newFilename := filepath.Join(toPath, contentFilename)
	err := os.Rename(oFilename, newFilename)
	if err != nil {
		return "", err
	}
	return newFilename, nil
}

func readFileName(rsp *resty.Response) string {
	content := rsp.Header().Get(contentDisposition)
	if content == "" {
		return ""
	}
	fields := strings.Split(content, ";")
	for _, field := range fields {
		trimField := strings.TrimSpace(field)
		if strings.HasPrefix(trimField, "filename=") {
			return strings.Split(trimField, "=")[1]
		}
	}
	return ""
}

func genTempFile() string {
	return uuid.New().String()
}
