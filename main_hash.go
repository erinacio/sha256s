package main

import (
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/minio/sha256-simd"
)

type hashResult struct {
	Name string
	Sum  []byte
}

func hashMain(opt Options) {
	nameCh := make(chan string)
	hashCh := make(chan hashResult)

	var hashWg sync.WaitGroup
	hashWg.Add(opt.Jobs)
	go func() {
		hashWg.Wait()
		close(hashCh)
	}()

	var errorsCount int64
	go walkWorker(opt, nameCh, &errorsCount)
	for range make([]struct{}, opt.Jobs) {
		go hashWorker(&hashWg, nameCh, hashCh, &errorsCount)
	}

	writer := HashSumWriter{
		Name:   "SHA256",
		Tag:    opt.Tag,
		Zero:   opt.Zero,
		Binary: opt.Binary,
	}
	if opt.NativePath {
		for result := range hashCh {
			writer.Write(os.Stdout, result.Sum, result.Name)
		}
	} else {
		for result := range hashCh {
			writer.Write(os.Stdout, result.Sum, toUnixPath(result.Name))
		}
	}

	if atomic.LoadInt64(&errorsCount) > 0 {
		os.Exit(1)
	}
}

func walkWorker(opt Options, nameCh chan<- string, errorsCount *int64) {
	defer close(nameCh)
	if opt.Recursive {
		statFn := os.Lstat
		if opt.Dereference {
			statFn = os.Stat
		}
		for _, path := range opt.Paths {
			FindRegularFiles(path, statFn, func(name string, err error) {
				if err == nil {
					nameCh <- name
				} else {
					atomic.AddInt64(errorsCount, 1)
					logError(err)
				}
			})
		}
	} else {
		for _, path := range opt.Paths {
			nameCh <- path
		}
	}
}

func hashWorker(wg *sync.WaitGroup, nameCh <-chan string, hashCh chan<- hashResult, errorsCount *int64) {
	defer wg.Done()
	hash := sha256.New()
	for name := range nameCh {
		sum, err := fileHash(hash, name)
		if err == nil {
			hashCh <- hashResult{
				Name: name,
				Sum:  sum,
			}
		} else {
			atomic.AddInt64(errorsCount, 1)
			logError(err)
		}
	}
}

func toUnixPath(nativePath string) (unixPath string) {
	if filepath.Separator == '/' {
		return nativePath
	} else {
		return strings.ReplaceAll(nativePath, string(filepath.Separator), "/")
	}
}
