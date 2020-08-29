package main

import (
	"fmt"
	"log"
	"os"
	"strings"
	"unicode/utf8"
)

var Version = "unknown"

func main() {
	log.SetFlags(0)
	log.SetPrefix("sha256s: ")

	var opt Options
	if err := opt.Parse(os.Args[1:]); err != nil && strings.HasSuffix(err.Error(), ": "+ErrHelpRequested.Error()) {
		opt.PrintHelp(os.Stdout)
		return
	} else if err != nil && strings.HasSuffix(err.Error(), ": "+ErrVersionRequested.Error()) {
		fmt.Printf("sha256s version %s\n", Version)
		return
	} else if err != nil {
		log.Printf("%s", err.Error())
		log.Printf("Try 'sha256s --help' for more information.")
		os.Exit(1)
	}

	if opt.Check {
		checkMain(opt)
	} else {
		hashMain(opt)
	}
}

func logError(err error) {
	if pe, ok := err.(*os.PathError); ok {
		errMsg := pe.Err.Error()
		_, size := utf8.DecodeRuneInString(errMsg)
		errMsg = strings.ToTitle(errMsg[:size]) + errMsg[size:]
		log.Printf("%s: %s", pe.Path, errMsg)
	} else {
		log.Printf("%v", err)
	}
}
