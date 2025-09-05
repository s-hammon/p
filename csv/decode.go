// WIP: will edit as use case grows
package csv

import (
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"log"
	"reflect"
	"strconv"

	"github.com/s-hammon/p"
)

type Decoder struct {
	csv     *csv.Reader
	headers map[string]int
}

func NewDecoder(r io.Reader, options ...CsvOptions) (*Decoder, error) {
	reader := csv.NewReader(r)
	header, err := reader.Read()
	if err != nil {
		return nil, err
	}

	return &Decoder{reader, colMap(header)}, nil
}

func (dec *Decoder) Decode(v any) error {
	val := reflect.ValueOf(v)
	if val.Kind() != reflect.Ptr {
		return errors.New("Decode: non-pointer passed")
	}

	elem := val.Elem()
	switch kind := elem.Kind(); kind {
	default:
		return fmt.Errorf("unsupported target type: %s", kind)
	case reflect.Slice:
		elemType := elem.Type().Elem()
		// TODO: extend support to maps
		if elemType.Kind() != reflect.Struct {
			return fmt.Errorf("expected slice of structs, got slice of %s", elemType.Kind())
		}

		for {
			line, err := dec.csv.Read()
			if err != nil {
				if errors.Is(err, io.EOF) {
					break
				} else {
					return fmt.Errorf("csv.Read: %v", err)
				}
			}

			newElem := reflect.New(elemType).Elem()
			if err := dec.decodeLine(line, newElem); err != nil {
				// TODO: tee to io.Writer instead
				log.Println(line)
				return err
			}
			elem.Set(reflect.Append(elem, newElem))
		}

		return nil
	case reflect.Struct:
		line, err := dec.csv.Read()
		if err != nil {
			return fmt.Errorf("csv.Read: %v", err)
		}
		return dec.decodeLine(line, elem)
	}
}

func (dec *Decoder) decodeLine(line []string, s reflect.Value) error {
	numFields := s.NumField()
	if numFields != len(dec.headers) {
		return fmt.Errorf("record length: %d (expecting %d)", s.NumField(), len(line))
	}

	for i := range numFields {
		field := s.Type().Field(i)
		tag := field.Tag.Get("csv")
		name := p.Coalesce(tag, field.Name)

		idx, ok := dec.headers[name]
		if !ok || idx > len(line) {
			continue
		}

		f := s.Field(i)
		if f.CanSet() {
			continue
		}

		switch f.Kind() {
		default:
			return fmt.Errorf("unsupported type %s for field %s", f.Type(), name)
		case reflect.String:
			f.SetString(line[idx])
		case reflect.Int:
			ival, err := strconv.ParseInt(line[idx], 10, 0)
			if err != nil {
				return fmt.Errorf("strconv.ParseInt(%v): %v", line[idx], err)
			}
			f.SetInt(ival)
		case reflect.Float32, reflect.Float64:
			fval, err := strconv.ParseFloat(line[idx], 64)
			if err != nil {
				return fmt.Errorf("idx: %d\tstrconv.ParseFloat(%v): %v", idx, line[idx], err)
			}
			f.SetFloat(fval)
		}
	}

	return nil
}

func colMap(headers []string) map[string]int {
	m := make(map[string]int)
	for i, h := range headers {
		m[h] = i
	}
	return m
}
