/*
Package search provides a common approach for capturing search queries.

As search queries are translated into database queries, there are implementations for
working within inmemory and MongoDB databases.
*/
package search

import (
	"fmt"
	"net/url"
	"strconv"
	"strings"

	"gitscm.cisco.com/ccdev/go-common/sets"
	"gitscm.cisco.com/mcmp/errors"

	"gitscm.cisco.com/mcmp/utils/ordered"
)

const (
	desc = "-"
)

// Option defines how to construct a search query.
type Option func(q *Query)

// Fields is an Option for specifying which field(s) should be returned during a search.
func Fields(field ...string) Option {
	return func(q *Query) {
		q.fields = append(q.fields, field...)
	}
}

// Limit is an Option for specifying max number of records should be returned during a search.
func Limit(n uint) Option {
	return func(q *Query) {
		q.limit = n
	}
}

// Offset is an Option for specifying where in records to start returning records; when results sorted.
func Offset(n uint) Option {
	return func(q *Query) {
		q.offset = n
	}
}

// Sortby is an Option for specifying how to sort the records found during a search.
func Sortby(sb ...string) Option {
	return func(q *Query) {
		q.convertSortBy(sb...)
	}
}

// Query holds the a search request.
type Query struct {
	// Count holds the number of results the query matches; used for pagination
	Count int64

	filters Filters
	fields  []string
	sortby  *ordered.Map
	limit   uint
	offset  uint
}

// NewQuery initializes a new Query to use as a search request.
func NewQuery(opts ...Option) *Query {
	q := &Query{
		filters: make(Filters),
		fields:  make([]string, 0),
		sortby:  ordered.NewMap(),
	}

	for _, opt := range opts {
		opt(q)
	}

	return q
}

// AddFilter inserts a filter to define filter criteria for a search.
func (q *Query) AddFilter(key string, value interface{}, op ...Operator) {
	q.filters.Set(newFilter(key, value, op...))
}

// RemoveFilter deletes a filter from the filters based on the specified key.
func (q *Query) RemoveFilter(key string) {
	q.filters.Remove(key)
}

// EmptyFields checks if query fields is empty.
func (q *Query) EmptyFields() bool {
	return len(q.fields) == 0
}

// EmptyFilters checks if the query filters is empty.
func (q *Query) EmptyFilters() bool {
	return q.filters.Len() == 0
}

// EmptySortby checks if the query sortby is empty.
func (q *Query) EmptySortby() bool {
	return q.sortby.Len() == 0
}

// Fields returns the query fields.
func (q *Query) Fields() []string {
	return q.fields
}

// Filters returns the query filters.
func (q *Query) Filters() Filters {
	return q.filters
}

// Limit returns the query limit.
func (q *Query) Limit() uint {
	return q.limit
}

// Offset returns the query offset.
func (q *Query) Offset() uint {
	return q.offset
}

// Sortby returns the query sortby ordered map.
func (q *Query) Sortby() *ordered.Map {
	return q.sortby
}

// Increment will increase the offset by the specified limit.
func (q *Query) Increment() {
	q.offset += q.limit
}

// Validate checks if the fields or sortby are valid attributes based on the set of attributes provided.
func (q *Query) Validate(attributes sets.String) error {
	if !q.EmptyFields() {
		fields := sets.NewString(q.fields...)
		if !attributes.HasAll(fields.Difference(attributes).List()...) {
			return errors.NewDomainError(errors.ErrInvalid, errors.Default, fmt.Sprintf("%v", fields.Difference(attributes).UnsortedList()), fmt.Sprintf("%v", attributes.UnsortedList()))
		}
	}

	if !q.EmptySortby() {
		iter := q.sortby.EntriesIter()

		for {
			pair, ok := iter()
			if !ok {
				break
			}

			if !attributes.Has(pair.Key) {
				return errors.NewDomainError(errors.ErrInvalid, errors.Default, pair.Key, fmt.Sprintf("a valid field name: %v", attributes.UnsortedList()))
			}
		}
	}

	if q.Offset() > 0 && q.EmptySortby() {
		return errors.NewDomainError(errors.ErrInvalid, errors.Default, fmt.Sprintf("%d", q.Offset()), "sortby not to be empty")
	}

	return nil
}

func (q *Query) String() string {
	v := make(url.Values)

	for k := range q.filters {
		if q.filters[k].Op == Equal {
			v.Add(q.filters[k].Key, fmt.Sprintf("%v", q.filters[k].Value))
		}
	}

	if !q.EmptyFields() {
		v.Add("fields", strings.Join(q.fields, ","))
	}

	if !q.EmptySortby() {
		var b strings.Builder

		iter := q.Sortby().EntriesIter()

		for {
			pair, ok := iter()
			if !ok {
				break
			}

			if !pair.Value.(bool) {
				_, _ = b.WriteString(desc)
			}

			_, _ = b.WriteString(pair.Key)
			_, _ = b.WriteString(",")
		}

		s := b.String()   // no copying
		s = s[:b.Len()-1] // no copying (removes trailing ",")

		v.Add("sort", s)
	}

	if q.limit > 0 {
		v.Add("limit", strconv.FormatInt(int64(q.limit), 10))
	}

	if q.offset > 0 {
		v.Add("offset", strconv.FormatInt(int64(q.offset), 10))
	}

	return v.Encode()
}

// convertSortBy converts a list of tagged fields into an ordered map indicating which fields are in ascending/descending order.
func (q *Query) convertSortBy(sb ...string) {
	for _, v := range sb {
		// if the field should be descending it will start with `-`
		q.sortby.Set(strings.TrimPrefix(v, desc), !strings.HasPrefix(v, desc))
	}
}
