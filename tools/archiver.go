package tools

import (
	"errors"
	"reflect"

	"github.com/mholt/archiver/v3"
)

var ErrUnsupported = errors.New("unsupported archiver")

func Unarchive(filename, path string) error {
	iua, err := archiver.ByExtension(filename)
	if err != nil {
		return ErrUnsupported
	}
	applyArchiverSettings(iua, 1)
	u, ok := iua.(archiver.Unarchiver)
	if ok {
		return u.Unarchive(filename, path)
	}
	return ErrUnsupported
}

func applyArchiverSettings(iua interface{}, stripComponent int64) {
	rua := reflect.ValueOf(iua).Elem()
	if v := rua.FieldByName("StripComponents"); v.CanSet() {
		v.SetInt(stripComponent)
	}
}
