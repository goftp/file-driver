package filedriver

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/goftp/server"
)

type FileDriver struct {
	RootPath string
	TmpDir   string
}

func (driver *FileDriver) ChangeDir(path string) error {
	rPath := filepath.Join(driver.RootPath, path)
	f, err := os.Lstat(rPath)
	if err != nil {
		return err
	}
	if f.IsDir() {
		return nil
	}
	return errors.New("Not a dir")
}

func (driver *FileDriver) Stat(path string) (os.FileInfo, error) {
	basepath := filepath.Join(driver.RootPath, path)
	rPath, err := filepath.Abs(basepath)
	if err != nil {
		return nil, err
	}
	return os.Lstat(rPath)
}

func (driver *FileDriver) DirContents(path string) ([]os.FileInfo, error) {
	files := make([]os.FileInfo, 0)
	basepath := filepath.Join(driver.RootPath, path)
	filepath.Walk(basepath, func(f string, info os.FileInfo, err error) error {
		rPath, _ := filepath.Rel(basepath, f)
		if rPath == info.Name() {
			files = append(files, info)
		}
		return nil
	})

	return files, nil
}

func (driver *FileDriver) DeleteDir(path string) error {
	rPath := filepath.Join(driver.RootPath, path)
	f, err := os.Lstat(rPath)
	if err != nil {
		return err
	}
	if f.IsDir() {
		return os.Remove(rPath)
	}
	return errors.New("Not a directory")
}

func (driver *FileDriver) DeleteFile(path string) error {
	rPath := filepath.Join(driver.RootPath, path)
	f, err := os.Lstat(rPath)
	if err != nil {
		return err
	}
	if !f.IsDir() {
		return os.Remove(rPath)
	}
	return errors.New("Not a file")
}

func (driver *FileDriver) Rename(fromPath string, toPath string) error {
	oldPath := filepath.Join(driver.RootPath, fromPath)
	newPath := filepath.Join(driver.RootPath, toPath)
	return os.Rename(oldPath, newPath)
}

func (driver *FileDriver) MakeDir(path string) error {
	rPath := filepath.Join(driver.RootPath, path)
	return os.Mkdir(rPath, os.ModePerm)
}

func (driver *FileDriver) GetFile(path string, offset int64) (int64, io.ReadCloser, error) {
	rPath := filepath.Join(driver.RootPath, path)
	f, err := os.Open(rPath)
	if err != nil {
		return 0, nil, err
	}

	info, err := f.Stat()
	if err != nil {
		return 0, nil, err
	}

	f.Seek(offset, os.SEEK_SET)

	return info.Size(), f, nil
}

func (driver *FileDriver) PutFile(destPath string, data io.Reader, appendData bool) (int64, error) {
	rPath := filepath.Join(driver.RootPath, destPath)
	var isExist bool
	f, err := os.Lstat(rPath)
	if err == nil {
		isExist = true
		if f.IsDir() {
			return 0, errors.New("A dir has the same name")
		}
	} else {
		if os.IsNotExist(err) {
			isExist = false
		} else {
			return 0, errors.New(fmt.Sprintln("Put File error:", err))
		}
	}

	fmt.Println("is append", appendData)

	if !appendData {
		if isExist {
			err = os.Remove(rPath)
			if err != nil {
				return 0, err
			}
		}
		f, err := os.Create(rPath)
		if err != nil {
			return 0, err
		}
		defer f.Close()
		bytes, err := io.Copy(f, data)
		if err != nil {
			return 0, err
		}
		return bytes, nil
	}

	if !isExist {
		return 0, errors.New("Append data but file not exsit")
	}

	of, err := os.OpenFile(rPath, os.O_APPEND|os.O_RDWR, 0660)
	if err != nil {
		return 0, err
	}
	defer of.Close()

	cur, err := of.Seek(0, os.SEEK_END)
	if err != nil {
		return 0, err
	}

	fmt.Println("cur pointer is", cur)

	bytes, err := io.Copy(of, data)
	if err != nil {
		return 0, err
	}

	return bytes, nil
}

type FileDriverFactory struct {
	RootPath string
	TmpDir   string
}

func (factory *FileDriverFactory) NewDriver() (server.Driver, error) {
	return &FileDriver{factory.RootPath, factory.TmpDir}, nil
}
