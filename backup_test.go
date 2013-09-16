package backup

import (
	"crypto/md5"
	crand "crypto/rand"
	"crypto/sha1"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"math/rand"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

func testBackup(storage Storage, t *testing.T) {
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
	if err != nil {
		t.Fatalf("%v", err)
	}
	err = Store(fileDir, metaFile, storage)
	if err != nil {
		t.Fatalf("%v", err)
	}
	fmt.Printf("stored\n")
	// retrieve
	retrieveDir := "test_retrieve"
	fmt.Printf("retrieving\n")
	err = deleteDir(retrieveDir)
	if err != nil {
		t.Fatalf("%v", err)
	}
	Walk(metaFile, storage, func(file *File) {
		err := file.Retrieve(retrieveDir, storage)
		if err != nil {
			t.Fatalf("%v", err)
		}
	})
	fmt.Printf("retrieved\n")
	// compare dir
	err = compareDir(fileDir, retrieveDir)
	if err != nil {
		t.Fatalf("%v", err)
	}
}

func compareDir(left, right string) error {
	lfs, err := ioutil.ReadDir(left)
	if err != nil {
		return err
	}
	rfs, err := ioutil.ReadDir(right)
	if err != nil {
		return err
	}
	if len(lfs) != len(rfs) {
		return errors.New("diff")
	}
	for i, linfo := range lfs {
		rinfo := rfs[i]
		lpath := filepath.Join(left, linfo.Name())
		rpath := filepath.Join(right, rinfo.Name())
		if linfo.IsDir() {
			// dir
			if !rinfo.IsDir() {
				return errors.New("diff")
			}
			compareDir(lpath, rpath)
		} else {
			// file
			lsig, err := getSig(lpath)
			if err != nil {
				return err
			}
			rsig, err := getSig(rpath)
			if err != nil {
				return err
			}
			if lsig != rsig {
				return errors.New("diff")
			}
		}
	}
	return nil
}

func getSig(filePath string) (string, error) {
	hash1 := md5.New()
	f, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	_, err = io.Copy(hash1, f)
	if err != nil {
		return "", err
	}
	f.Close()
	hash2 := sha1.New()
	f, err = os.Open(filePath)
	if err != nil {
		return "", err
	}
	_, err = io.Copy(hash2, f)
	if err != nil {
		return "", err
	}
	f.Close()
	return fmt.Sprintf("%x%x", hash1.Sum(nil), hash2.Sum(nil)), nil
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
