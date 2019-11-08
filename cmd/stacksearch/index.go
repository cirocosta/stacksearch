package main

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/cirocosta/stacksearch/pkg"
)

type indexCommand struct {
	Paths []string `long:"profile" short:"p"`

	// Once indicates that we should just "index" in-memory.
	//
	Once bool `long:"once" description:"in-memory search for a single search"`
}

func (c *indexCommand) Execute(args []string) (err error) {
	dataset, err := populateDataset(c.Paths)
	if err != nil {
		return
	}

	var callstacks []pkg.Callstack
	for _, argument := range args {
		callstacks, err = dataset.Get(argument)
		if err != nil {
			return
		}
	}

	for _, callstack := range callstacks {
		fmt.Println(strings.Join(callstack.Data, "\n"))
		fmt.Println()
	}

	return
}

func populateDataset(paths []string) (dataset pkg.Dataset, err error) {
	var matches []string

	for _, path := range paths {
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

	dataset = pkg.NewMemory()
	for _, callstack := range callstacks {
		dataset.Add(callstack)
	}

	return
}
