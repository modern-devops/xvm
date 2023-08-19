package tools

import (
	"errors"
	"reflect"

	"github.com/mholt/archiver/v3"
)

var errorUnsupported = errors.New("unsupported archiver")

// Unarchive 解包
func Unarchive(filename, toPath string) error {
	return archive(filename, toPath, 1)
}

// UnarchiveRoot 解压根目录下的所有文件
func UnarchiveRoot(filename, toPath string) error {
	return archive(filename, toPath, 0)
}

func archive(filename, toPath string, stripComponent int64) error {
	iua, err := archiver.ByExtension(filename)
	if err != nil {
		return errorUnsupported
	}
	patchArchiverSettings(iua, stripComponent)
	u, ok := iua.(archiver.Unarchiver)
	if ok {
		return u.Unarchive(filename, toPath)
	}
	return errorUnsupported
}

func patchArchiverSettings(iua interface{}, stripComponent int64) {
	rua := reflect.ValueOf(iua).Elem()
	v := rua.FieldByName("StripComponents")
	if v.CanSet() {
		v.SetInt(stripComponent)
	}
}
