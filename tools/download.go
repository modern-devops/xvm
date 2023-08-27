package tools

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/google/uuid"
	"github.com/schollz/progressbar/v3"
)

const contentDisposition = "Content-Disposition"

func Download(url string, path string) (string, error) {
	fp := filepath.Join(path, getFilename(url))
	resp, err := download(url, fp)
	if err != nil {
		return "", err
	}
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("Failed to download, status: %d", resp.StatusCode)
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

func download(url, filename string) (*http.Response, error) {
	resp, _ := http.Get(url)
	defer resp.Body.Close()

	if err := os.MkdirAll(filepath.Dir(filename), os.ModeDir); err != nil {
		return nil, err
	}
	f, _ := os.OpenFile(filename, os.O_CREATE|os.O_WRONLY, 0644)
	defer f.Close()

	bar := progressbar.DefaultBytes(
		resp.ContentLength,
		"downloading",
	)
	_, err := io.Copy(io.MultiWriter(f, bar), resp.Body)
	return resp, err
}

func rename(resp *http.Response, fp string) (string, error) {
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

func readFileName(rsp *http.Response) string {
	content := rsp.Header.Get(contentDisposition)
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
