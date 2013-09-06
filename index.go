package backup

import (
	"bytes"
	"crypto/md5"
	"crypto/sha1"
	"encoding/gob"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"time"
)

const BLOB_LENGTH = 32 * 1024 * 1024
const FILEREAD_BUFFER_SIZE = 8 * 1024 * 1024

type File struct {
	ModTime time.Time
	Path    string
	Size    int64
	Blobs   map[int64]*Blob
}

type Blob struct {
	Offset int64
	Length int64
	Hash   string
}

func IndexDir(topDir string, metaFilepath string) {
	buf := make([]byte, FILEREAD_BUFFER_SIZE)
	files := make(map[string]*File)
	metaFile, err := os.Open(metaFilepath)
	if err == nil {
		gob.NewDecoder(metaFile).Decode(&files)
	}
	fmt.Printf("read %d files from meta file\n", len(files))
	endWriter, waitWriter := startWriter(files, metaFilepath)
	err = Walk(topDir, func(path string, fileinfo os.FileInfo, err error) error {
		// check error
		if err != nil {
			log.Fatal(err)
		}
		// skip hidden files and dirs
		if path != "." && path[0] == '.' {
			if fileinfo.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}
		if fileinfo.IsDir() {
			return nil
		}
		fmt.Printf("indexing %s\n", path)
		// file struct
		file, ok := files[path]
		if !ok {
			file = &File{
				ModTime: fileinfo.ModTime(),
				Path:    path,
				Size:    fileinfo.Size(),
				Blobs:   make(map[int64]*Blob),
			}
		} else {
			// check file change
			if file.ModTime != fileinfo.ModTime() || file.Size != fileinfo.Size() {
				// all blob is dirty
				file.ModTime = fileinfo.ModTime()
				file.Size = fileinfo.Size()
				file.Blobs = make(map[int64]*Blob)
			}
		}
		// calculate offset
		var offset int64
		for {
			if blob, ok := file.Blobs[offset]; ok {
				offset += blob.Length
			} else {
				break
			}
		}
		fmt.Printf("start at offset %d\n", offset)
		if offset == fileinfo.Size() { // all indexed
			fmt.Printf("all indexed\n")
			return nil
		}
		// index
		f, err := os.Open(filepath.Join(topDir, path))
		f.Seek(int64(offset), 0)
		hasher := md5.New()
		hasher2 := sha1.New()
		var length, start int64
		for {
			n, err := f.Read(buf)
			if n > 0 {
				hasher.Write(buf[:n])
				hasher2.Write(buf[:n])
				offset += int64(n)
				length += int64(n)
				if length >= BLOB_LENGTH {
					sum := fmt.Sprintf("%x%x", hasher.Sum(nil), hasher2.Sum(nil))
					file.Blobs[start] = &Blob{
						Offset: start,
						Length: length,
						Hash:   sum,
					}
					fmt.Printf("new blob: %d %d %s\n", start, length, sum)
					length = 0
					start = offset
					hasher.Reset()
					hasher2.Reset()
				}
			}
			if err == io.EOF {
				sum := fmt.Sprintf("%x%x", hasher.Sum(nil), hasher2.Sum(nil))
				file.Blobs[start] = &Blob{
					Offset: start,
					Length: length,
					Hash:   sum,
				}
				fmt.Printf("new blob: %d %d %s\n", start, length, sum)
				break
			} else if err != nil {
				log.Fatalf("error when reading %s: %v", path, err)
			}
		}
		// save
		files[path] = file
		return nil
	})
	if err != nil {
		log.Fatal(err)
	}
	endWriter <- true
	<-waitWriter
}

func startWriter(files map[string]*File, metaFilepath string) (chan bool, chan bool) {
	end := make(chan bool)
	wait := make(chan bool)
	go func() {
		ticker := time.NewTicker(time.Second * 5)
	loop:
		for {
			select {
			case <-ticker.C:
				writeMeta(files, metaFilepath)
				continue loop
			case <-end:
				writeMeta(files, metaFilepath)
				wait <- true
				break loop
			}
		}
	}()
	return end, wait
}

func writeMeta(files map[string]*File, metaFilepath string) {
	out, err := os.Create(metaFilepath)
	if err != nil {
		log.Fatal(err)
	}
	buf := new(bytes.Buffer)
	gob.NewEncoder(buf).Encode(files)
	out.Write(buf.Bytes())
	out.Close()
}
