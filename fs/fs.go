package fs

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/cloudacademy/s3zipper/core"
)

type Filesystem struct{}

func New() *Filesystem {
	return &Filesystem{}
}

func (f *Filesystem) GetReader(path string) (rdr io.ReadCloser, err error) {
	rdr, err = os.Open(path)
	if err != nil {
		err = fmt.Errorf("Error opening file \"%s\" - %s", path, err.Error())
	}
	return
}

func (f *Filesystem) List(prefix string) (files []core.ZipItem, err error) {
	paths := []string{}

	err = filepath.Walk(prefix, func(path string, f os.FileInfo, err error) error {
		if f.IsDir() == false {
			paths = append(paths, path)
		}
		return nil
	})

	if err != nil {
		return
	}
	for _, path := range paths {
		key := strings.TrimPrefix(path, prefix)
		filename := filepath.Base(key)
		folder := filepath.Dir(key)
		files = append(files, FSZipItem{
			FileName: filename,
			Folder:   folder,
			FullPath: path,
		})
	}
	return
}

func (f *Filesystem) CacheExists(prefix string) (bool, error) {
	return false, nil
}

func (f *Filesystem) CacheSignedUrl(prefix string) string {
	return "#"
}

type FSZipItem struct {
	FileName string
	Folder   string
	FullPath string
}

func (s FSZipItem) GetFilename() string {
	return s.FileName
}
func (s FSZipItem) GetFolder() string {
	return s.Folder
}
func (s FSZipItem) GetPath() string {
	return s.FullPath
}
