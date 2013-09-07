package backup

import (
	crand "crypto/rand"
	"fmt"
	"io"
	"math/rand"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

func TestBackup(t *testing.T) {
	// generate files
	fileDir := "test_data"
	deleteDir(fileDir)
	err := os.Mkdir(fileDir, 0755)
	if err != nil {
		t.Fatalf("mkdir %v", err)
	}
	fmt.Printf("generating")
	err = generateRandomFileOrDirs(fileDir, 3)
	if err != nil {
		t.Fatalf("gen dir %v", err)
	}
	fmt.Printf("\ngenerated\n")
	// index
	fmt.Printf("indexing\n")
	metaFile := "test_meta"
	os.Remove(metaFile)
	err = IndexDir("test_data", metaFile, 512*1024)
	if err != nil {
		t.Fatalf("%v", err)
	}
	fmt.Printf("indexed\n")
	// store
	fmt.Printf("storing\n")
	deleteDir("test_store")
	storage, err := NewLocalStorage("test_store")
	if err != nil {
		t.Fatalf("%v", err)
	}
	err = Store(fileDir, metaFile, storage)
	if err != nil {
		t.Fatalf("%v", err)
	}
	fmt.Printf("stored\n")
	// retrieve
	fmt.Printf("retrieving\n")
	err = deleteDir("test_retrieve")
	if err != nil {
		t.Fatalf("%v", err)
	}
	Walk(metaFile, storage, func(file *File) {
		err := file.Retrieve("test_retrieve", storage)
		if err != nil {
			t.Fatalf("%v", err)
		}
	})
	fmt.Printf("retrieved\n")
	//TODO compare dir
}

func generateRandomFileOrDirs(dir string, depth int) error {
	if depth == 0 {
		return nil
	}
	n := 4 + rand.Intn(4)
	for i := 0; i < n; i++ {
		nameBytes := make([]byte, 32)
		for i := range nameBytes {
			nameBytes[i] = byte(rand.Int31() & 0xff)
		}
		name := fmt.Sprintf("%x", nameBytes)
		path := filepath.Join(dir, name)
		print(".")
		os.Stdout.Sync()
		if rand.Intn(2) == 0 { // dir
			os.Mkdir(path, 0755)
			generateRandomFileOrDirs(path, depth-1)
		} else { // file
			f, err := os.Create(path)
			if err != nil {
				return err
			}
			io.CopyN(f, crand.Reader, int64(rand.Intn(8*1024*1024)))
			f.Close()
		}
	}
	return nil
}

func deleteDir(dir string) error {
	f, err := os.Open(dir)
	if err != nil {
		return err
	}
	subs, err := f.Readdir(-1)
	if err != nil {
		return err
	}
	for _, sub := range subs {
		if sub.IsDir() {
			err := deleteDir(filepath.Join(dir, sub.Name()))
			if err != nil {
				return err
			}
		} else {
			err := os.Remove(filepath.Join(dir, sub.Name()))
			if err != nil {
				return err
			}
		}
	}
	err = os.Remove(dir)
	if err != nil {
		return err
	}
	return nil
}
