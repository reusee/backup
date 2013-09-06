package backup

import (
	"testing"
)

func TestStore(t *testing.T) {
	storage, err := NewLocalStorage("store_anime")
	if err != nil {
		t.Fatal(err)
	}
	Store("/media/anime", "meta", storage)
}
