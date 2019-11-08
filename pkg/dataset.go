package pkg

type Dataset interface {
	// Add adds a new stack to the dataset.
	//
	Add(stack Callstack) (err error)

	// Get retrieves the set of callstacks where `fn` is present.
	//
	Get(fn string) (callstacks []Callstack, err error)

	// Funcs retrieves the set of available functions for querying.
	//
	Funcs() (fns []string)
}

type memory struct {
	kv map[string][]Callstack
}

func NewMemory() *memory {
	return &memory{
		kv: map[string][]Callstack{},
	}
}

func (d *memory) Add(stack Callstack) (err error) {
	for _, method := range stack.Data {
		_, found := d.kv[method]
		if found {
			d.kv[method] = append(d.kv[method], stack)
			continue
		}

		d.kv[method] = []Callstack{stack}
	}

	return
}

func (d memory) Funcs() (fns []string) {
	fns = make([]string, len(d.kv))

	i := 0
	for k := range d.kv {
		fns[i] = k
		i += 1
	}

	return
}

func (d memory) Get(fn string) (callstacks []Callstack, err error) {
	var found bool

	callstacks, found = d.kv[fn]
	if !found {
		return
	}

	return
}
