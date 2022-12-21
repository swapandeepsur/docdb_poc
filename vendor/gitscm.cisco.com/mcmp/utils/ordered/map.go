// Package ordered provides a type Map for use in JSON handling
// although JSON spec says the keys order of an object should not matter
// but sometimes when working with particular third-party proprietary code
// which has incorrect using the keys order, we have to maintain the object keys
// in the same order of incoming JSON object, this package is useful for these cases.
//
package ordered

// Refers
//  JSON and Go        https://blog.golang.org/json-and-go
//  Go-Ordered-JSON    https://github.com/virtuald/go-ordered-json
//  Python OrderedDict https://github.com/python/cpython/blob/2.7/Lib/collections.py#L38
//  port OrderedDict   https://github.com/cevaris/ordered_map

import (
	"bytes"
	"container/list"
	"encoding/json"
	"fmt"
	"io"
	"reflect"
	"sync"

	wraperrors "github.com/pkg/errors"
)

// KVPair holds the key-value pair type, for initializing from a list of key-value pairs, or for looping entries in the same order.
type KVPair struct {
	Key   string
	Value interface{}
}

// Map type, has similar operations as the default map, but maintained
// the keys order of inserted; similar to map, all single key operations (Get/Set/Delete) runs at O(1).
type Map struct {
	mu   sync.RWMutex
	m    map[string]interface{}
	l    *list.List
	keys map[string]*list.Element // the double linked list for delete and lookup to be O(1)
}

// NewMap creates a new Map.
func NewMap() *Map {
	return &Map{
		m:    make(map[string]interface{}),
		l:    list.New(),
		keys: make(map[string]*list.Element),
	}
}

// NewMapFromKVPairs creates a new Map and populate from a list of key-value pairs.
func NewMapFromKVPairs(pairs []*KVPair) *Map {
	om := NewMap()

	for _, pair := range pairs {
		om.Set(pair.Key, pair.Value)
	}

	return om
}

// Set value for particular key, this will remember the order of keys inserted
// but if the key already exists, the order is not updated.
func (om *Map) Set(key string, value interface{}) {
	om.mu.Lock()
	defer om.mu.Unlock()

	if _, ok := om.m[key]; !ok {
		om.keys[key] = om.l.PushBack(key)
	}

	om.m[key] = value
}

// Has check if value exists.
func (om *Map) Has(key string) bool {
	om.mu.RLock()
	defer om.mu.RUnlock()

	_, ok := om.m[key]

	return ok
}

// Get value for particular key, or nil if not exist; but don't rely on nil for non-exist; should check by Has or GetValue.
func (om *Map) Get(key string) (value interface{}) {
	om.mu.RLock()
	defer om.mu.RUnlock()

	value = om.m[key]

	return
}

// GetValue returns value and exists together.
func (om *Map) GetValue(key string) (value interface{}, ok bool) {
	om.mu.RLock()
	defer om.mu.RUnlock()

	value, ok = om.m[key]

	return
}

// Delete the element with the specified key (m[key]) from the map. If there is no such element, this is a no-op.
func (om *Map) Delete(key string) (value interface{}, ok bool) {
	om.mu.Lock()
	defer om.mu.Unlock()

	value, ok = om.m[key]
	if ok {
		om.l.Remove(om.keys[key])
		delete(om.keys, key)
		delete(om.m, key)
	}

	return
}

// Len returns the size of the map/number of elements.
func (om *Map) Len() int {
	om.mu.RLock()
	defer om.mu.RUnlock()

	return om.l.Len()
}

// Equal returns true
// 	- if both Map are nil
// 	- both are non-nil, same length with same key-value pair values in the same order
// Otherwise returns false.
func (om *Map) Equal(other *Map) bool {
	if om == nil && other == nil {
		return true
	}

	if !(om != nil && other != nil) {
		return false
	}

	if om.Len() != other.Len() {
		return false
	}

	iterA := om.EntriesIter()
	iterB := other.EntriesIter()

	for {
		pairA, okA := iterA()
		if !okA {
			break
		}

		pairB, _ := iterB()

		if !pairA.Equal(pairB) {
			return false
		}
	}

	return true
}

// EntriesIter will iterate all key/value pairs in the same order of object constructed.
func (om *Map) EntriesIter() func() (*KVPair, bool) {
	e := om.l.Front()

	return func() (*KVPair, bool) {
		if e != nil {
			key := e.Value.(string)
			e = e.Next()

			return &KVPair{Key: key, Value: om.m[key]}, true
		}

		return nil, false
	}
}

// EntriesReverseIter will iterate all key/value pairs in the reverse order of object constructed.
func (om *Map) EntriesReverseIter() func() (*KVPair, bool) {
	e := om.l.Back()

	return func() (*KVPair, bool) {
		if e != nil {
			key := e.Value.(string)
			e = e.Prev()

			return &KVPair{Key: key, Value: om.m[key]}, true
		}

		return nil, false
	}
}

// MarshalJSON implements type json.Marshaler interface, so can be called in json.Marshal(om).
func (om *Map) MarshalJSON() (res []byte, err error) {
	om.mu.RLock()
	defer om.mu.RUnlock()

	res = append(res, '{')
	front, back := om.l.Front(), om.l.Back()

	for e := front; e != nil; e = e.Next() {
		k := e.Value.(string)
		res = append(res, fmt.Sprintf("%q:", k)...)

		var b []byte

		b, err = json.Marshal(om.m[k])
		if err != nil {
			return
		}

		res = append(res, b...)

		if e != back {
			res = append(res, ',')
		}
	}

	res = append(res, '}')

	return
}

// UnmarshalJSON implements type json.Unmarshaler interface, so can be called in json.Unmarshal(data, om).
func (om *Map) UnmarshalJSON(data []byte) error {
	dec := json.NewDecoder(bytes.NewReader(data))
	dec.UseNumber()

	// must open with a delim token '{'
	t, err := dec.Token()
	if err != nil {
		return err
	}

	if delim, ok := t.(json.Delim); !ok || delim != '{' {
		return wraperrors.Errorf("expect JSON object open with '{'")
	}

	if err = om.parseobject(dec); err != nil {
		return err
	}

	t, err = dec.Token()
	if err != io.EOF {
		return wraperrors.Errorf("expect end of JSON object but got more token: %T: %v or err: %v", t, t, err)
	}

	return nil
}

func (om *Map) parseobject(dec *json.Decoder) (err error) {
	var t json.Token

	for dec.More() {
		t, err = dec.Token()
		if err != nil {
			return err
		}

		key, ok := t.(string)
		if !ok {
			return wraperrors.Errorf("expecting JSON key should be always a string: %T: %v", t, t)
		}

		t, err = dec.Token()
		if err == io.EOF {
			break
		} else if err != nil {
			return err
		}

		var value interface{}

		value, err = handledelim(t, dec)
		if err != nil {
			return err
		}

		om.Set(key, value)
	}

	t, err = dec.Token()
	if err != nil {
		return err
	}

	if delim, ok := t.(json.Delim); !ok || delim != '}' {
		return wraperrors.Errorf("expect JSON object close with '}'")
	}

	return nil
}

// Equal returns true if both KVPair are nil or both non-nil and Key and Value are equal
// otherwise returns false.
func (p *KVPair) Equal(other *KVPair) bool {
	if p == nil && other == nil {
		return true
	}

	if !(p != nil && other != nil) {
		return false
	}

	return p.Key == other.Key && reflect.DeepEqual(p.Value, other.Value)
}

func parsearray(dec *json.Decoder) (arr []interface{}, err error) {
	var t json.Token

	arr = make([]interface{}, 0)

	for dec.More() {
		t, err = dec.Token()
		if err != nil {
			return arr, err
		}

		var value interface{}

		value, err = handledelim(t, dec)
		if err != nil {
			return arr, err
		}

		arr = append(arr, value)
	}

	t, err = dec.Token()
	if err != nil {
		return arr, err
	}

	if delim, ok := t.(json.Delim); !ok || delim != ']' {
		err = wraperrors.Errorf("expect JSON array close with ']'")

		return arr, err
	}

	return arr, err
}

func handledelim(t json.Token, dec *json.Decoder) (res interface{}, err error) {
	if delim, ok := t.(json.Delim); ok {
		switch delim {
		case '{':
			om := NewMap()
			if err = om.parseobject(dec); err != nil {
				return
			}

			return om, nil
		case '[':
			var value []interface{}

			value, err = parsearray(dec)
			if err != nil {
				return
			}

			return value, nil
		default:
			return nil, wraperrors.Errorf("unexpected delimiter: %q", delim)
		}
	}

	return t, nil
}
