package pkg

import (
	"crypto/sha256"
	"fmt"
	"math"
	"os"
	"regexp"
	"runtime"
	"strings"

	pprof "github.com/google/pprof/profile"
	"github.com/pkg/errors"
)

type Location struct {
	Line     int64
	Filename string
}

type Callstack struct {
	Data      []string
	Locations []Location

	digest [32]byte
}

type CallstackOptions struct {
	StopAt  *regexp.Regexp
	Verbose bool
}

type CallstackOption func(*CallstackOptions)

func WithStopAt(matcher *regexp.Regexp) CallstackOption {
	return func(o *CallstackOptions) {
		o.StopAt = matcher
	}
}

func WithVerbose(v bool) CallstackOption {
	return func(o *CallstackOptions) {
		o.Verbose = v
	}
}

func NewCallstack(data []string, locations []Location, opt *CallstackOptions) (callstack Callstack) {
	callstack.Data = data
	callstack.Locations = locations

	if opt != nil {
		if opt.StopAt != nil {
			idx := 0
			for _, dp := range data {
				if !opt.StopAt.MatchString(dp) {
					idx++
					continue
				}

				callstack.Data = callstack.Data[idx:]
				break
			}
		}
	}

	callstack.digest = sha256.Sum256([]byte(
		strings.Join(callstack.Data, "\n")),
	)

	return
}

func isEqual(a, b []string) bool {
	for _, vA := range a {
		for _, vB := range b {
			if vA != vB {
				return false
			}
		}
	}

	return true
}

func subCallstackMatch(a, b Callstack) []Callstack {
	lA, lB := len(a.Data), len(b.Data)

	switch {
	case lA == lB:
		if isEqual(a.Data, b.Data) {
			return []Callstack{a}
		}
	case lA < lB:
		if isEqual(a.Data, b.Data[:lA]) {
			return []Callstack{b}
		}
	case lA > lB:
		if isEqual(a.Data[:lB], b.Data) {
			return []Callstack{a}
		}
	}

	return []Callstack{a, b}
}

type callstackSet struct {
	stacks []Callstack
	kv     map[[32]byte]struct{}
}

func newSet() callstackSet {
	return callstackSet{
		kv: map[[32]byte]struct{}{},
	}
}

func (c *callstackSet) add(callstacks ...Callstack) {
	for _, callstack := range callstacks {
		_, found := c.kv[callstack.digest]
		if found {
			return
		}

		c.kv[callstack.digest] = struct{}{}
		c.stacks = append(c.stacks, callstack)
	}
}

func MergeSubCallstacks(callstacks []Callstack) (merged []Callstack) {
	s := newSet()

	for i := 0; i < len(callstacks)-1; i++ {
		for j := i + 1; j < len(callstacks); j++ {
			s.add(subCallstackMatch(
				callstacks[i], callstacks[j],
			)...)
		}
	}

	merged = s.stacks

	return
}

func loadFileStacks(opts CallstackOptions, files <-chan string, results chan<- []Callstack, errsC chan<- error) {
	for file := range files {
		profile, err := loadPprofProfile(file)
		if err != nil {
			errsC <- fmt.Errorf(
				"failed to load pprof profile %s: %w", file, err,
			)
			return
		}

		rawStacks, err := CallstacksFromPprof(profile, opts)
		if err != nil {
			errsC <- fmt.Errorf(
				"failed to convert from pprof to internal format: %w",
				err,
			)
			return
		}

		results <- rawStacks
	}
}

// LoadCallstacks retrieves the unique set of callstacks across all profiles.
//
func LoadCallstacks(files []string, opts ...CallstackOption) (callstacks []Callstack, err error) {
	var (
		cOpts   = CallstackOptions{}
		stacksC = make(chan []Callstack)
		filesC  = make(chan string, len(files))
		errsC   = make(chan error, len(files))
	)

	for _, opt := range opts {
		opt(&cOpts)
	}

	poolSize := math.Min(float64(runtime.NumCPU()), float64(len(files)))
	for i := 0; i < int(poolSize); i++ {
		go loadFileStacks(cOpts, filesC, stacksC, errsC)
	}

	for _, file := range files {
		filesC <- file
	}
	close(filesC)

	var (
		allStacks      []Callstack
		filesCollected int
	)

wait:
	for {
		select {
		case stacks := <-stacksC:
			allStacks = append(allStacks, stacks...)
			filesCollected++
			if filesCollected == len(files) {
				break wait
			}
		case err = <-errsC:
			return
		}

	}

	callstacks = Merge(allStacks)
	return
}

// sampleStack captures the callstack of a given sample.
//
func sampleStack(sample *pprof.Sample, opts *CallstackOptions) (callstack Callstack) {
	var (
		data      = make([]string, len(sample.Location))
		locations []Location
	)

	if opts.Verbose {
		locations = make([]Location, len(sample.Location))
	}

	for idx, location := range sample.Location {
		data[idx] = location.Line[0].Function.Name

		if opts.Verbose {
			locations[idx] = Location{
				Filename: location.Line[0].Function.Filename,
				Line:     location.Line[0].Line,
			}
		}
	}

	callstack = NewCallstack(data, locations, opts)

	return
}

// FromPprof converts a pprof profile to a set of unique stacks.
//
func CallstacksFromPprof(src *pprof.Profile, opts CallstackOptions) (callstacks []Callstack, err error) {
	if src == nil {
		err = errors.Errorf("src profile must no be nil")
		return
	}

	s := newSet()

	for _, sample := range src.Sample {
		callstack := sampleStack(sample, &opts)
		s.add(callstack)
	}

	callstacks = s.stacks

	return
}

// Merge takes two sets of callstacks and produce a final one that has only
// unique callstacks.
//
func Merge(callstacks []Callstack) []Callstack {
	s := newSet()

	s.add(callstacks...)
	return s.stacks
}

// loadPprofProfile loads a *.pprof profile from disk into an in-memory parsed
// format.
//
func loadPprofProfile(file string) (profile *pprof.Profile, err error) {
	f, err := os.Open(file)
	if err != nil {
		err = fmt.Errorf("failed to read profile file %s: %w", file, err)
		return
	}

	defer f.Close()

	profile, err = pprof.Parse(f)
	if err != nil {
		err = fmt.Errorf("failed parsing profile from file %s: %w",
			file, err)
		return
	}

	return
}
