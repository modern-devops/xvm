package linker

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"

	"github.com/modern-devops/xvm/tools/commander"
)

const fileMode = 0744

type Option int

const (
	None           Option = iota
	OverrideAlways        = 1 << 0
)

// New register a command who call as `name` under directory bin
// example: New('test', '/usr/local/bin', 'echo')
func New(name, bin, command string, options ...Option) (string, error) {
	err := os.MkdirAll(bin, fileMode)
	if err != nil {
		return "", fmt.Errorf("failed to create directory: %s, %w", bin, err)
	}
	option := applyOptions(options)
	if runtime.GOOS != "windows" {
		return linkToUnixLike(name, bin, command, option)
	}
	binPath, err := linkToWin32(name, bin, command, option)
	if err != nil {
		return "", err
	}
	// msys2
	return binPath, linkToMsys2(name, bin, command, option)
}

func applyOptions(options []Option) Option {
	option := None
	for _, o := range options {
		option |= o
	}
	return option
}

func linkToWin32(name, binPath, command string, options Option) (string, error) {
	f := cmdFile(binPath, name)
	if skipLink(f, options) {
		return f, nil
	}
	return f, ioutil.WriteFile(f, cmdTemplate(command), fileMode)
}

// linkToMsys2 win10 git bash
func linkToMsys2(name, binPath, command string, options Option) error {
	f := msys2File(binPath, name)
	if skipLink(f, options) {
		return nil
	}
	args := commander.Parse(command)
	if len(args) == 0 {
		return fmt.Errorf("empty args: %s", command)
	}
	execuableFile := args[0]
	if _, err := os.Stat(execuableFile); err == nil {
		unixCommandPath := toUnixLikePath(execuableFile)
		command = strings.Replace(command, execuableFile, unixCommandPath, 1)
	}
	return os.WriteFile(f, shellTemplate(command), fileMode)
}

func linkToUnixLike(name, bin, command string, options Option) (string, error) {
	file := filepath.Join(bin, name)
	if skipLink(file, options) {
		return file, nil
	}
	template := shellTemplate(command)
	if err := os.WriteFile(file, template, fileMode); err != nil {
		return "", fmt.Errorf("failed to create file: %s, %w", file, err)
	}
	return file, nil
}

func skipLink(file string, options Option) bool {
	if _, err := os.Stat(file); err == nil && !hasOption(options, OverrideAlways) {
		return true
	}
	return false
}

// ToUnixLikePath translate windows style path to unix like style path
// C:\Users\Administrator\Go\src\bin\gofmt.exe
// ===>
// /c/Users/Administrator/Go/src/bin/gofmt.exe
func toUnixLikePath(path string) string {
	path = strings.ReplaceAll(path, "\\", "/")
	if match, err := regexp.MatchString(`^[A-Za-z]:/.*`, path); err == nil && match {
		// translate C:/Users/... to /c/Users/...
		path = "/" + strings.ToLower(path[0:1]) + path[2:]
	}
	return path
}

func cmdFile(binPath, name string) string {
	return filepath.Join(binPath, fmt.Sprintf(`%s.cmd`, name))
}

func cmdTemplate(command string) []byte {
	return []byte(fmt.Sprintf("@echo off\n\n%s %%*", command))
}

func shellTemplate(command string) []byte {
	return []byte(fmt.Sprintf("#!/bin/sh\n\n%s \"$@\"", command))
}

func msys2File(binPath, name string) string {
	return filepath.Join(binPath, name)
}

func hasOption(option Option, targetOption Option) bool {
	return option&targetOption == targetOption
}
