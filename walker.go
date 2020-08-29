package main

import (
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
)

type WalkFunc func(name string, err error)

type statFunc func(name string) (info os.FileInfo, err error)

func FindRegularFiles(path string, statFn statFunc, walkFn WalkFunc) {
	info, err := os.Stat(path) // intentionally
	if err != nil {
		walkFn(path, err)
		return
	}
	if mode := info.Mode(); mode.IsRegular() {
		walkFn(path, nil)
	} else if mode.IsDir() {
		findRegularFilesInDir(path, statFn, walkFn)
	}
}

func findRegularFilesInDir(root string, statFn statFunc, walkFn WalkFunc) {
	file, err := os.Open(root)
	if err != nil {
		walkFn(root, err)
		return
	}
	defer file.Close()

	for {
		names, err := file.Readdirnames(128)
		if err != nil {
			if err != io.EOF {
				walkFn(root, err)
			}
			return
		}
		for _, name := range names {
			path := filepath.Join(root, name)
			looping, err := isLooping(path)
			if err != nil {
				walkFn(path, err)
				continue
			}
			if looping {
				walkFn(path, &os.PathError{Op: "stat", Path: path, Err: errors.New("symlink loop detected")})
				continue
			}
			info, err := statFn(path)
			if err != nil {
				walkFn(path, err)
				continue
			}
			if mode := info.Mode(); mode.IsRegular() {
				walkFn(path, nil)
			} else if mode.IsDir() {
				findRegularFilesInDir(path, statFn, walkFn)
			}
		}
	}
}

func isLooping(path string) (looping bool, err error) {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return
	}
	visitedSet := make(map[string]struct{}, strings.Count(path, string(filepath.Separator)))
	for {
		var realPath, file string
		realPath, err = filepath.EvalSymlinks(absPath)
		if err != nil {
			return
		}
		_, looping = visitedSet[realPath]
		if looping {
			return
		}
		visitedSet[realPath] = struct{}{}
		absPath, file = filepath.Split(absPath)
		if file == "" {
			return
		}
		absPath = filepath.Clean(absPath)
	}
}
