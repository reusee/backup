package backup

import (
	"io"
	"os"
	"path/filepath"
)

type LocalStorage struct {
	path string
}

func NewLocalStorage(path string) (*LocalStorage, error) {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return nil, err
	}
	if _, err := os.Stat(absPath); err != nil {
		if os.IsNotExist(err) { // create
			os.Mkdir(absPath, 0755)
		} else {
			return nil, err
		}
	}
	ret := &LocalStorage{
		path: absPath,
	}
	return ret, nil
}

func (self *LocalStorage) Set(key string, data []byte) error {
	path := filepath.Join(self.path, key)
	if _, err := os.Stat(path); err == nil { // already exists
		return nil
	}
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	f.Write(data)
	f.Close()
	return nil
}

func (self *LocalStorage) Get(key string, writer io.Writer) error {
	path := filepath.Join(self.path, key)
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()
	io.Copy(writer, f)
	return nil
}
