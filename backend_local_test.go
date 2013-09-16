package backup

import (
	"testing"
)

func TestLocal(t *testing.T) {
	deleteDir("test_store")
	storage, err := NewLocalStorage("test_store")
	if err != nil {
		t.Fatalf("%v", err)
	}
	testBackup(storage, t)
}
