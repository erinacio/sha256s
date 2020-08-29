package main

import (
	"io"
	"io/ioutil"
	"os"
	"strings"
	"sync/atomic"
)

var stdinOpened uint32

func OpenFile(path string) (file io.ReadCloser, err error) {
	if path != "-" {
		return os.Open(path)
	}
	if atomic.CompareAndSwapUint32(&stdinOpened, 0, 1) {
		return os.Stdin, nil
	} else {
		return ioutil.NopCloser(strings.NewReader("")), nil
	}
}
