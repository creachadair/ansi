// Program ansi generates ANSI escape sequences to stdout.
package main

import (
	"flag"
	"io"
	"log"
	"os"

	"bitbucket.org/creachadair/ansi/ansi"
)

// TODO(fromberger): Other useful things.

var setTitle = flag.String("title", "", "Set terminal title (OSC 0)")

func main() {
	flag.Parse()
	if *setTitle == "" {
		log.Fatal("You must specify a -title to set")
	}

	c := ansi.NewCoder(os.Stdout)
	io.WriteString(c.SetIf(']', "0;", 0, "\007"), *setTitle)
}
