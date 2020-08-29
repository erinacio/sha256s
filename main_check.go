package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/minio/sha256-simd"
)

type checksumLine struct {
	ArgI int
	Name string
	Sum  []byte
}

type checkResult struct {
	ArgI int
	Name string
	Stat string
}

func checkMain(opt Options) {
	if opt.Status {
		log.SetOutput(ioutil.Discard)
		_ = os.Stdout.Close()
	}
	lineCh := make(chan checksumLine)
	checkCh := make(chan checkResult)

	var checkWg sync.WaitGroup
	checkWg.Add(opt.Jobs)
	go func() {
		checkWg.Wait()
		close(checkCh)
	}()

	var badLinesCount, badFilesCount, errorsCount, mismatchCount int64
	go readSumWorker(opt, lineCh, &badLinesCount, &errorsCount)
	for range make([]struct{}, opt.Jobs) {
		go checkWorker(&checkWg, opt, lineCh, checkCh, &badFilesCount)
	}

	noFileVerifiedSet, fileVerifiedSet := make(map[int]struct{}), make(map[int]struct{}, len(opt.Paths))
	for result := range checkCh {
		if result.Stat == "" {
			if _, ok := fileVerifiedSet[result.ArgI]; !ok {
				noFileVerifiedSet[result.ArgI] = struct{}{}
			}
			continue
		}
		if _, ok := noFileVerifiedSet[result.ArgI]; ok {
			delete(noFileVerifiedSet, result.ArgI)
			fileVerifiedSet[result.ArgI] = struct{}{}
		}
		if strings.IndexByte(result.Name, '\n') >= 0 {
			result.Name, _ = escapeName(result.Name)
			result.Name = `\` + result.Name
		}
		if result.Stat == "OK" {
			if !opt.Quiet {
				fmt.Printf("%s: %s\n", result.Name, result.Stat)
			}
		} else if result.Stat == "FAILED" {
			mismatchCount++
			fmt.Printf("%s: %s\n", result.Name, result.Stat)
		} else {
			fmt.Printf("%s: %s\n", result.Name, result.Stat)
		}
	}

	for i := range noFileVerifiedSet {
		atomic.AddInt64(&errorsCount, 1)
		log.Printf("%s: no file was verified", opt.Paths[i])
	}
	if c := atomic.LoadInt64(&badLinesCount); c > 0 {
		log.Printf("WARNING: %d %s improperly formatted", c, iif(c == 1, "line is", "lines are"))
	}
	if c := atomic.LoadInt64(&badFilesCount); c > 0 {
		log.Printf("WARNING: %d listed %s could not be read", c, iif(c == 1, "file", "files"))
	}
	if c := mismatchCount; c > 0 {
		log.Printf("WARNING: %d computed %s did not match", c, iif(c == 1, "checksum", "checksums"))
	}
	if (opt.Strict && atomic.LoadInt64(&badLinesCount) != 0) ||
		atomic.LoadInt64(&badFilesCount) != 0 ||
		atomic.LoadInt64(&errorsCount) != 0 ||
		mismatchCount != 0 {
		os.Exit(1)
	}
}

func iif(b bool, t, f string) string {
	if b {
		return t
	} else {
		return f
	}
}

func readSumWorker(opt Options, lineCh chan<- checksumLine, badLinesCount *int64, errorsCount *int64) {
	defer close(lineCh)
	reader := HashSumReader{
		Name:  "SHA256",
		Width: sha256.Size,
		Tag:   opt.Tag,
		Zero:  opt.Zero,
		CrLf:  opt.CrLf,
	}
	for i, path := range opt.Paths {
		reader.Read(path, func(sum []byte, name string, err error) {
			if err == nil {
				lineCh <- checksumLine{
					ArgI: i,
					Name: name,
					Sum:  sum,
				}
			} else if _, ok := err.(BadLineError); ok {
				if opt.Warn {
					logError(err)
				}
				atomic.AddInt64(badLinesCount, 1)
			} else {
				logError(err)
				atomic.AddInt64(errorsCount, 1)
			}
		})
	}
}

func checkWorker(wg *sync.WaitGroup, opt Options, lineCh <-chan checksumLine, checkCh chan<- checkResult, badFilesCount *int64) {
	defer wg.Done()
	hash := sha256.New()
	for line := range lineCh {
		var name string
		if opt.NativePath {
			name = line.Name
		} else {
			name = fromUnixPath(name)
		}
		sum, err := fileHash(hash, line.Name)
		if opt.IgnoreMissing && os.IsNotExist(err) {
			checkCh <- checkResult{
				ArgI: line.ArgI,
				Name: line.Name,
			}
		} else if err != nil {
			logError(err)
			checkCh <- checkResult{
				ArgI: line.ArgI,
				Name: line.Name,
				Stat: "FAILED open or read",
			}
			atomic.AddInt64(badFilesCount, 1)
		} else if bytes.Equal(sum, line.Sum) {
			checkCh <- checkResult{
				ArgI: line.ArgI,
				Name: line.Name,
				Stat: "OK",
			}
		} else {
			checkCh <- checkResult{
				ArgI: line.ArgI,
				Name: line.Name,
				Stat: "FAILED",
			}
		}
	}
}

func fromUnixPath(nativePath string) (unixPath string) {
	if filepath.Separator == '/' {
		return nativePath
	} else {
		return strings.ReplaceAll(nativePath, "/", string(filepath.Separator))
	}
}
