package main

import (
	"hash"
	"io"
)

func fileHash(hash hash.Hash, name string) (sum []byte, err error) {
	file, err := OpenFile(name)
	if err != nil {
		return
	}
	defer file.Close()

	hash.Reset()
	if _, err = io.Copy(hash, file); err == nil {
		sum = hash.Sum(nil)
	}
	return
}
