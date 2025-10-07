package p

import (
	"bufio"
	"crypto/md5"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"reflect"
	"slices"
	"strings"
)

func Marshal(v any) string {
	ret, _ := json.Marshal(v)
	return string(ret)
}

func Tee(r io.Reader, limit int) io.Reader {
	pipeIn, pipeOut := io.Pipe()
	count := 0
	go func() {
		defer pipeOut.Close()

		scanner := bufio.NewScanner(r)
		for scanner.Scan() {
			count++
			if count > limit {
				break
			}
			bytes := scanner.Bytes()
			fmt.Println(string(bytes))
			pipeOut.Write(append(bytes, []byte("\n")...))
		}
	}()

	return pipeIn
}

func Format(tmpl string, elems ...any) string {
	return fmt.Sprintf(tmpl, elems...)
}

// Coalesce will return the first non-zero value in elems.
func Coalesce[T any](elems ...T) T {
	var res T
	for len(elems) > 0 {
		if res = elems[0]; !IsZero(res) {
			return res
		}
		elems = elems[1:]
	}
	return res
}

// If evaluates the stmt, and returns either the true or false result.
// What can I say? Can't you tell I come from a SQL/data background?
func If[T any](stmt bool, trueRes, falseRes T) T {
	if stmt {
		return trueRes
	}

	return falseRes
}

// convert map to list of key-value pairs ("key=value")
// useful for injecting params in URL payload
func SerializeMap[T comparable](m map[string]T) []string {
	props := make([]string, 0, len(m))
	for k, v := range m {
		props = append(props, Format("%v=%v", k, v))
	}
	return props
}

// convert map to key-value pairs where the value is quoted (key="value")
// useful for things like adding props to HTML tags
func SerializeMapQuoteVal[T comparable](m map[string]T) []string {
	props := make([]string, 0, len(m))
	for k, v := range m {
		props = append(props, Format(`%v="%v"`, k, v))
	}
	return props
}

// deserializes []string (see `SerializeMap`) to map[string]string
// requires both key and value to be non-zero (and later not null)
// TODO: support for `map[string]any`
func StringDeserialize(props []string) map[string]string {
	m := make(map[string]string)
	for _, prop := range props {
		parts := strings.SplitN(prop, "=", 2)
		if len(Set(parts)) == 2 {
			m[parts[0]] = parts[1]
		}
	}
	return m
}

// filters an array with the filter function 'fn'
func Filter[T any](ss []T, fn func(T) bool) (ret []T) {
	for _, s := range ss {
		if fn(s) {
			ret = append(ret, s)
		}
	}
	return ret
}

func IsIn[T comparable](v T, s ...T) bool {
	return slices.Contains(s, v)
}

// returns the unique set of (hashable) non-zero elements in a slice/array
func Set[T comparable](ss []T) []T {
	set := SetM(ss)
	return Keys(set)
}

// same as above, but returns as a map
func SetM[T comparable](ss []T) map[T]struct{} {
	set := make(map[T]struct{})
	for _, s := range ss {
		if IsZero(s) {
			continue
		}
		set[s] = struct{}{}
	}
	return set
}

// returns the set of all elements between two slices
func Union[S ~[]E, E comparable](s1, s2 S) S {
	res := S{}
	res = append(res, s1...)
	for _, v := range s2 {
		if !slices.Contains(res, v) {
			res = append(res, v)
		}
	}
	return res
}

type node struct {
	children map[rune]*node
	isEnd    bool
}

func newNode() *node {
	return &node{
		children: make(map[rune]*node),
	}
}

type Trie struct {
	root *node
}

func NewTrie() *Trie {
	return &Trie{
		root: newNode(),
	}
}

func (t *Trie) Search(s string) bool {
	current := t.root
	for _, c := range s {
		if _, ok := current.children[c]; !ok {
			return false
		}
		current = current.children[c]
	}
	return current.isEnd
}

func (t *Trie) Add(s string) {
	current := t.root
	for _, c := range s {
		if _, ok := current.children[c]; !ok {
			current.children[c] = newNode()
		}
		current = current.children[c]
	}
	current.isEnd = true
}

// makes a trie from a list of strings
func (t *Trie) AddMany(elements ...string) {
	for _, element := range elements {
		t.Add(element)
	}
}

// returns the set of common elements between two slices
func Intersection[T comparable](s1, s2 []T) []T {
	set := SetM(s1)
	res := []T{}
	for _, v := range s2 {
		if _, ok := set[v]; ok {
			res = append(res, v)
		}
	}
	return res
}

func Keys[K comparable, V any](m map[K]V) []K {
	ret := make([]K, 0, len(m))
	for k := range m {
		ret = append(ret, k)
	}
	return ret
}

// checks if val is a zero value of its underlying type (0, "", false, etc)
func IsZero(val any) bool {
	return reflect.ValueOf(val).IsZero()
}

func IsNil(val any) bool {
	if val == nil {
		return true
	}
	switch t := val.(type) {
	case map[any]any:
		if t == nil {
			return true
		}
	case []any:
		if len(t) == 0 {
			return true
		}
	}
	return false
}

func SHA256(data ...string) string {
	hash := sha256.Sum256([]byte(strings.Join(data, "")))
	return hex.EncodeToString(hash[:])
}

func MD5(data ...string) string {
	hash := md5.Sum([]byte(strings.Join(data, "")))
	return hex.EncodeToString(hash[:])
}
