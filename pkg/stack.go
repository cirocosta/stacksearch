package pkg

import (
	"crypto/sha256"
	"fmt"
	"os"
	"strings"

	pprof "github.com/google/pprof/profile"
	"github.com/pkg/errors"
)

// Callstack represents a sequence of function calls.
//
type Callstack struct {
	Data   []string
	digest [32]byte
}

func NewCallstack(data []string) (callstack Callstack) {
	callstack.Data = data
	callstack.digest = sha256.Sum256([]byte(
		strings.Join(callstack.Data, "\n")),
	)

	return
}

// LoadCallstacks retrieves the unique set of callstacks across all profiles.
//
func LoadCallstacks(files []string) (callstacks []Callstack, err error) {
	var (
		profile   *pprof.Profile
		rawStacks []Callstack
	)

	allCallstacks := []Callstack{}

	// map
	for _, file := range files {
		profile, err = loadPprofProfile(file)
		if err != nil {
			err = fmt.Errorf("failed to load pprof profile %s: %w", file, err)
			return
		}

		rawStacks, err = CallstacksFromPprof(profile)
		if err != nil {
			err = fmt.Errorf("failed to convert from pprof to internal format: %w",
				err)
			return
		}

		allCallstacks = append(allCallstacks, rawStacks...)
	}

	callstacks = Merge(allCallstacks)

	return
}

// SampleStack captures the callstack of a given sample.
//
func SampleStack(sample *pprof.Sample) (callstack Callstack) {
	data := make([]string, len(sample.Location))

	for idx, location := range sample.Location {
		data[idx] = location.Line[0].Function.Name
	}

	callstack = NewCallstack(data)

	return
}

// HavingFn filters down the list of stacks to those containing a function.
//
func HavingFn(stacks []Callstack, fn string) (res []Callstack, err error) {
	return
}

// FromPprof converts a pprof profile to a set of unique stacks.
//
func CallstacksFromPprof(src *pprof.Profile) (callstacks []Callstack, err error) {
	if src == nil {
		err = errors.Errorf("src profile must no be nil")
		return
	}

	m := map[[32]byte]struct{}{}

	for _, sample := range src.Sample {
		callstack := SampleStack(sample)

		_, found := m[callstack.digest]
		if found {
			continue
		}

		m[callstack.digest] = struct{}{}
		callstacks = append(callstacks, callstack)
	}

	return
}

// Merge takes two sets of callstacks and produce a final one that has only
// unique callstacks.
//
func Merge(callstacks []Callstack) (merged []Callstack) {
	m := map[[32]byte]struct{}{}

	for _, callstack := range callstacks {
		_, found := m[callstack.digest]
		if found {
			continue
		}

		m[callstack.digest] = struct{}{}
		merged = append(merged, callstack)
	}

	return
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
