package backup

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type ListCallback func(file *File)

func Walk(metaFilepath string, storage Storage, cb ListCallback) {
	files := readMetaFile(metaFilepath)
	for _, file := range files {
		var offset int64
		for {
			if blob, ok := file.Blobs[offset]; ok {
				offset += blob.Length
			} else {
				break
			}
		}
		if offset == file.Size { // complete file
			if cb != nil {
				cb(file)
			}
		}
	}
}

func (self *File) Retrieve(dir string, storage Storage) error {
	os.Mkdir(dir, 0755)
	dirs := strings.Split(filepath.Dir(self.Path), string(filepath.Separator))
	for i := 1; i <= len(dirs); i++ {
		path := filepath.Join(dir, filepath.Join(dirs[0:i]...))
		os.Mkdir(path, 0755)
	}
	f, err := os.Create(filepath.Join(dir, self.Path))
	if err != nil {
		return err
	}
	defer f.Close()
	fmt.Printf("retriving %s\n", self.Path)
	//TODO check hash before retrive
	var offset int64
	for {
		if blob, ok := self.Blobs[offset]; ok {
			key := fmt.Sprintf("%s-%d", blob.Hash, blob.Length)
			err := storage.Get(key, f)
			if err != nil {
				return err
			}
			offset += blob.Length
		} else {
			break
		}
	}
	if offset != self.Size {
		return errors.New("file not complete")
	}
	return nil
}
