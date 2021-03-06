package backup

import (
	"os"
	"path/filepath"
)

type walkfunc func(string, os.FileInfo, error) error

func walk(topDir string, cb walkfunc) error {
	info, err := os.Stat(topDir)
	if err != nil {
		return cb(topDir, nil, err)
	}
	baseDir, _ := filepath.Abs(topDir)
	return walk_(baseDir, topDir, info, cb)
}

func walk_(baseDir, path string, info os.FileInfo, cb walkfunc) error {
	absPath, _ := filepath.Abs(path)
	relativePath, _ := filepath.Rel(baseDir, absPath)
	err := cb(relativePath, info, nil)
	if err != nil {
		if info.IsDir() && err == filepath.SkipDir {
			return nil
		}
		return err
	}
	if !info.IsDir() {
		return nil
	}
	f, err := os.Open(path)
	if err != nil {
		return cb(path, info, err)
	}
	list, err := f.Readdir(-1)
	f.Close()
	if err != nil {
		return cb(path, info, err)
	}
	for _, fileInfo := range list {
		err = walk_(baseDir, filepath.Join(path, fileInfo.Name()), fileInfo, cb)
		if err != nil {
			if !fileInfo.IsDir() || err != filepath.SkipDir {
				return err
			}
		}
	}
	return nil
}
