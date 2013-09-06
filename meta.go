package backup

import (
	"encoding/gob"
	"fmt"
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
	fmt.Printf("load %d files from meta file\n", len(files))
	return files
}
