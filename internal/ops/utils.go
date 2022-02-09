package ops

import (
	"errors"
	"log"
	"os"
)

func exists(filename string) bool {
	_, err := os.Stat(filename)
	if err == nil {
		return true
	}

	if errors.Is(err, os.ErrNotExist) {
		return false
	}

	log.Fatalln("Unknown error opening the file: \n", err)
	return false

}
