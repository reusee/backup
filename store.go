package backup

import (
	"crypto/md5"
	"crypto/sha1"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
)

func Store(dir string, metaFilepath string, storage Storage) {
	files := readMetaFile(metaFilepath)
	hasher1 := md5.New()
	hasher2 := sha1.New()
	for path, file := range files {
		fmt.Printf("%s\n", path)
		f, err := os.Open(filepath.Join(dir, path))
		if err != nil {
			log.Fatal(err)
		}
		for _, blob := range file.Blobs {
			//TODO fast skip
			f.Seek(blob.Offset, 0)
			buf := make([]byte, blob.Length)
			n, err := io.ReadFull(f, buf)
			if int64(n) != blob.Length {
				log.Fatal(err)
			}
			hasher1.Write(buf)
			hasher2.Write(buf)
			sum := fmt.Sprintf("%x%x", hasher1.Sum(nil), hasher2.Sum(nil))
			if sum != blob.Hash {
				log.Fatal("hash not match") // TODO reindex
			}
			hasher1.Reset()
			hasher2.Reset()
			err = storage.Set(fmt.Sprintf("%s-%d", sum, blob.Length), buf)
			if err != nil {
				log.Fatal(err)
			}
		}
		f.Close()
	}
}

type Storage interface {
	Set(key string, data []byte) error
	Get(key string, writer io.Writer) error
}
