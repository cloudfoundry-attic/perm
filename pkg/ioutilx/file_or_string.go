package ioutilx

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"
)

var (
	OS       = InjectableOS{}
	IOReader = InjectableIOReader{}
)

type FileOrString string

func (f FileOrString) Bytes(statter Statter, reader FileReader) ([]byte, error) {
	value := string(f)
	stat, err := statter.Stat(value)
	if err != nil {
		return []byte(strings.Replace(value, "\\n", "\n", -1)), nil
	}

	if stat.IsDir() {
		return nil, fmt.Errorf("path '%s' is a directory, not a file", value)
	}

	return reader.ReadFile(value)
}

//go:generate counterfeiter . FileReader

type FileReader interface {
	ReadFile(string) ([]byte, error)
}

type InjectableIOReader struct{}

func (InjectableIOReader) ReadFile(filename string) ([]byte, error) {
	return ioutil.ReadFile(filename)
}

//go:generate counterfeiter os.FileInfo
//go:generate counterfeiter . Statter

type Statter interface {
	Stat(string) (os.FileInfo, error)
}

type InjectableOS struct{}

func (InjectableOS) Stat(name string) (os.FileInfo, error) {
	return os.Stat(name)
}
