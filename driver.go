package filedriver

import (
	"io"
	"os"
	"path/filepath"

	"github.com/go-xweb/log"
	"github.com/goftp/server"
)

type FileDriver struct {
	RootPath string
}

func (driver *FileDriver) ChangeDir(path string) bool {
	rPath := filepath.Join(driver.RootPath, path)
	f, err := os.Lstat(rPath)
	if err != nil {
		return false
	}
	return f.IsDir()
}

func (driver *FileDriver) Stat(path string) (os.FileInfo, error) {
	basepath := filepath.Join(driver.RootPath, path)
	return os.Stat(basepath)
}

func (driver *FileDriver) DirContents(path string) []os.FileInfo {
	files := make([]os.FileInfo, 0)
	basepath := filepath.Join(driver.RootPath, path)
	filepath.Walk(basepath, func(f string, info os.FileInfo, err error) error {
		rPath, _ := filepath.Rel(basepath, f)
		if rPath == info.Name() {
			files = append(files, info)
		}
		return nil
	})

	return files
}

func (driver *FileDriver) DeleteDir(path string) bool {
	rPath := filepath.Join(driver.RootPath, path)
	f, err := os.Lstat(rPath)
	if err != nil {
		return false
	}
	if f.IsDir() {
		os.Remove(rPath)
		return true
	}
	return false
}

func (driver *FileDriver) DeleteFile(path string) bool {
	rPath := filepath.Join(driver.RootPath, path)
	f, err := os.Lstat(rPath)
	if err != nil {
		return false
	}
	if !f.IsDir() {
		os.Remove(rPath)
		return true
	}
	return false
}

func (driver *FileDriver) Rename(fromPath string, toPath string) bool {
	oldPath := filepath.Join(driver.RootPath, fromPath)
	newPath := filepath.Join(driver.RootPath, toPath)
	err := os.Rename(oldPath, newPath)
	if err != nil {
		log.Errorf("rename %v to %v error: %v", fromPath, toPath, err)
	}
	return err == nil
}

func (driver *FileDriver) MakeDir(path string) bool {
	rPath := filepath.Join(driver.RootPath, path)
	err := os.Mkdir(rPath, os.ModePerm)
	if err != nil {
		log.Errorf("make dir %v error: %v", path, err)
	}
	return err == nil
}

func (driver *FileDriver) GetFile(path string) (int64, io.ReadCloser, error) {
	rPath := filepath.Join(driver.RootPath, path)
	f, err := os.Open(rPath)
	if err != nil {
		return 0, nil, err
	}

	info, err := f.Stat()
	if err != nil {
		return 0, nil, err
	}

	return info.Size(), f, nil
}

func (driver *FileDriver) PutFile(destPath string, data io.Reader) error {
	rPath := filepath.Join(driver.RootPath, destPath)
	f, err := os.OpenFile(rPath, os.O_RDWR|os.O_CREATE|os.O_TRUNC, os.ModePerm)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = io.Copy(f, data)
	return err
}

type FileDriverFactory struct {
	RootPath string
}

func (factory *FileDriverFactory) NewDriver() (server.Driver, error) {
	return &FileDriver{factory.RootPath}, nil
}
