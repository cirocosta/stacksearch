package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/cirocosta/stacksearch/pkg"
	"github.com/jessevdk/go-flags"
)

type cli struct {
	Paths     []string `long:"profile" required:"true" short:"p"`
	Verbose   bool     `long:"verbose" short:"v"`
	ShowFuncs bool     `long:"show-funcs" description:"shows available functions"`
}

func main() {
	var opts cli

	args, err := flags.NewParser(&opts, flags.HelpFlag|flags.PassDoubleDash).Parse()
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}

	err = run(opts, args)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
}

func run(c cli, args []string) (err error) {
	var callstacks []pkg.Callstack

	dataset, err := populateDataset(c.Paths)
	if err != nil {
		return
	}

	for _, argument := range args {
		callstacks, err = dataset.Get(argument)
		if err != nil {
			return
		}
	}

	if c.ShowFuncs {
		for _, f := range dataset.Funcs() {
			fmt.Println(f)
		}
		return
	}

	showCallstacks(callstacks, c.Verbose)

	return
}

func showCallstacks(callstacks []pkg.Callstack, verbose bool) {
	for _, callstack := range callstacks {
		for idx := range callstack.Data {
			fmt.Println(callstack.Data[idx])

			if !verbose {
				continue
			}

			fmt.Printf("\t%s:%d\n",
				callstack.Locations[idx].Filename,
				callstack.Locations[idx].Line,
			)
		}

		fmt.Println()
	}
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
