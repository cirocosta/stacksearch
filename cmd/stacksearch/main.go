package main

import (
	"fmt"
	"os"

	"github.com/jessevdk/go-flags"
)

var stacksearch struct {
	Index indexCommand `command:"index"`
}

func main() {

	parser := flags.NewParser(&stacksearch, flags.HelpFlag|flags.PassDoubleDash)
	parser.NamespaceDelimiter = "-"

	_, err := parser.Parse()
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
}
