package backup

import (
	"testing"
)

func TestStore(t *testing.T) {
	storage, err := NewLocalStorage("test_store")
	if err != nil {
		t.Fatal(err)
	}
	Store("/media/anime", "meta", storage)
}
