package backup

import (
	"encoding/gob"
	"log"
	"os"
)

func readMetaFile(metaFilepath string) map[string]*File {
	files := make(map[string]*File)
	f, err := os.Open(metaFilepath)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()
	gob.NewDecoder(f).Decode(&files)
	return files
}
