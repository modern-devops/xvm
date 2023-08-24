package tools

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-resty/resty/v2"
	"github.com/google/uuid"
)

const contentDisposition = "Content-Disposition"

func Download(url string, path string) (string, error) {
	fp := filepath.Join(path, getFilename(url))
	resp, err := download(url, fp)
	if err != nil {
		return "", err
	}
	if resp.IsError() {
		return "", fmt.Errorf("failed to download: %s", resp.String())
	}
	return rename(resp, fp)
}

func getFilename(url string) string {
	lastSlash := strings.LastIndex(url, "/")
	if lastSlash == -1 {
		return genTempFile()
	}
	filename := url[lastSlash+1:]
	firstQuestion := strings.Index(filename, "?")
	if firstQuestion != -1 {
		filename = filename[0:firstQuestion]
	}
	return filename
}

func download(url, filename string) (*resty.Response, error) {
	return resty.New().R().SetOutput(filename).Get(url)
}

func rename(resp *resty.Response, fp string) (string, error) {
	name := readFileName(resp)
	if name == "" {
		return fp, nil
	}
	tmpName := filepath.Base(fp)
	if name == tmpName {
		return fp, nil
	}
	newPath := filepath.Join(filepath.Dir(fp), name)
	if err := os.Rename(fp, newPath); err != nil {
		return "", err
	}
	return newPath, nil
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
