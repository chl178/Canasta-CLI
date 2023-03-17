package logging

import (
	"log"
)

var (
	verbose bool
)

func SetVerbose(v bool) {
	verbose = v
}

func Print(output string) {
	log.SetFlags(0)
	if verbose {
		log.Print(output)
	}
}

func Fatal(err error) {
	log.SetFlags(0)
	log.Fatal(err)
}
