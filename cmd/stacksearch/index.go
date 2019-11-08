package main

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/cirocosta/stacksearch/pkg"
)

type indexCommand struct {
	Paths []string `long:"profile" short:"p"`
}

func (c *indexCommand) Execute(args []string) (err error) {
	var (
		paths   = []string{}
		matches []string
	)

	for _, path := range c.Paths {
		matches, err = filepath.Glob(path)
		if err != nil {
			return
		}

		paths = append(paths, matches...)
	}

	callstacks, err := pkg.LoadCallstacks(paths)
	if err != nil {
		return
	}

	for _, callstack := range callstacks {
		fmt.Println(strings.Join(callstack.Data, "\n"))
		fmt.Println()
	}

	return
}
