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

func Coalesce[T any](a, b T) T {
	if IsZero(a) {
		return b
	}
	return a
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

// returns the unique set of (hashable) elements in a slice/array
func Set[T comparable](ss []T) []T {
	set := SetM(ss)
	return Keys(set)
}

// same as above, but returns as a map
func SetM[T comparable](ss []T) map[T]struct{} {
	set := make(map[T]struct{})
	for _, s := range ss {
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
