package search

// Operator represents different filter operations.
type Operator int

// Defined Operator values.
const (
	Equal Operator = iota
	LTE
	GTE
	NotEqual
	Like
	IgnoreCase
)

// Filter defines a search filter.
type Filter struct {
	Key   string
	Value interface{}
	Op    Operator
}

func newFilter(key string, value interface{}, op ...Operator) Filter {
	f := Filter{
		Key:   key,
		Value: value,
		Op:    Equal,
	}

	if len(op) > 0 {
		f.Op = op[0]
	}

	return f
}

// IgnoreCase returns true if the operator associated with the Filter is IgnoreCase.
func (f Filter) IgnoreCase() bool {
	return f.Op == IgnoreCase
}

// Like returns true if the operator associated with the Filter is Like.
func (f Filter) Like() bool {
	return f.Op == Like
}

// NotEqual returns true if the operator associated with the Filter is NotEqual.
func (f Filter) NotEqual() bool {
	return f.Op == NotEqual
}

// GTE returns true if the operator associated with the Filter is GTE.
func (f Filter) GTE() bool {
	return f.Op == GTE
}

// LTE returns true if the operator associated with the Filter is LTE.
func (f Filter) LTE() bool {
	return f.Op == LTE
}

// Filters defines a list of search filters.
type Filters map[string]Filter

// Set adds or overwrites a filter in the list of search filters.
func (f Filters) Set(v Filter) {
	f[v.Key] = v
}

// Remove will delete a filter based on the specified key.
func (f Filters) Remove(key string) {
	delete(f, key)
}

// Has checks if the specified key exists within the list of search filters.
func (f Filters) Has(key string) bool {
	_, ok := f[key]

	return ok
}

// Len is the number of elements in the collection of search filters.
func (f Filters) Len() int {
	return len(f)
}
