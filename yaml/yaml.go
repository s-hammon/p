// NOTE: WIP
package yaml

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"reflect"
	"strconv"
	"strings"

	"github.com/s-hammon/p"
)

type state int

const nl byte = 0x0A // newline, '\n'

const (
	stateStart state = iota + 1
	stateDocStart
	stateKey
	stateVal
	stateListItem
	stateScalar
	stateNewLine
	stateLiteralBlock
	stateComment
	stateDocEnd
	stateErr
	stateEnd
)

type Decoder struct {
	r      *bufio.Reader
	parser *parser
}

func NewDecoder(r io.Reader) *Decoder {
	return &Decoder{
		r:      bufio.NewReader(r),
		parser: newParser(),
	}
}

func (d *Decoder) Decode(val any) error {
	v := reflect.ValueOf(val)
	if v.Kind() != reflect.Ptr || v.Elem().Kind() != reflect.Map {
		return fmt.Errorf("Decode: expected pointer to map[string]any, got %T", val)
	}

	// Initialize map if nil
	if v.Elem().IsNil() {
		v.Elem().Set(reflect.MakeMap(v.Elem().Type()))
	}

	root := v.Elem()
	currentMap := root

	mapStack := p.NewStack[reflect.Value]()
	mapStack.Push(root)

	// Track the "current key" for list support
	var lastKey string

	for {
		line, err := d.r.ReadString('\n')
		if err != nil && !errors.Is(err, io.EOF) {
			return err
		}

		line = strings.TrimRight(line, "\r\n")
		if strings.TrimSpace(line) == "" || strings.HasPrefix(strings.TrimSpace(line), "#") {
			if errors.Is(err, io.EOF) {
				break
			}
			continue
		}

		indent := len(line) - len(strings.TrimLeft(line, " "))
		d.parser.checkIndent(indent)
		if d.parser.state == stateErr {
			return fmt.Errorf("invalid indentation at line: %q", line)
		}

		// Adjust nesting
		for mapStack.Len() > d.parser.indentStack.Len() {
			_, _ = mapStack.Pop()
		}
		currentMap, _ = mapStack.Peek()

		trimmed := strings.TrimSpace(line)

		// --- Handle map entries (key: value) ---
		if strings.Contains(trimmed, ":") {
			parts := strings.SplitN(trimmed, ":", 2)
			key := strings.TrimSpace(parts[0])
			value := strings.TrimSpace(parts[1])

			lastKey = key

			if value == "" {
				// Nested map
				nested := reflect.MakeMap(v.Elem().Type())
				currentMap.SetMapIndex(reflect.ValueOf(key), nested)
				mapStack.Push(nested)
			} else {
				currentMap.SetMapIndex(reflect.ValueOf(key), reflect.ValueOf(parseScalar(value)))
			}

			// --- Handle list items (- value) ---
		} else if strings.HasPrefix(trimmed, "- ") {
			if lastKey == "" {
				return fmt.Errorf("list item without a parent key: %q", trimmed)
			}

			item := strings.TrimSpace(strings.TrimPrefix(trimmed, "- "))

			existing := currentMap.MapIndex(reflect.ValueOf(lastKey))
			var arr []any
			if existing.IsValid() {
				arr = existing.Interface().([]any)
			}
			arr = append(arr, parseScalar(item))
			currentMap.SetMapIndex(reflect.ValueOf(lastKey), reflect.ValueOf(arr))

		} else {
			return fmt.Errorf("unexpected line format: %q", trimmed)
		}

		if errors.Is(err, io.EOF) {
			break
		}
	}

	return nil
}

func parseScalar(s string) any {
	s = strings.TrimSpace(s)

	switch strings.ToLower(s) {
	case "true":
		return true
	case "false":
		return false
	case "null", "~":
		return nil
	}

	if i, err := strconv.Atoi(s); err == nil {
		return i
	}

	if f, err := strconv.ParseFloat(s, 64); err != nil {
		return f
	}

	return s
}
