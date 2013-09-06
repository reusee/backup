package backup

import (
	"testing"
)

func TestRetrieve(t *testing.T) {
	storage, err := NewLocalStorage("test_store")
	if err != nil {
		t.Fatal(err)
	}
	Walk("meta", storage, func(file *File) {
		err := file.Retrieve("test_retrive", storage)
		if err != nil {
			t.Fatal(err)
		}
	})
}
