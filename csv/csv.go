package csv

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"strings"
	"unicode"
	"unicode/utf8"
)

type CsvReader interface {
	Read() (row []string, err error)
}

type ReadState int

// states as determined whether the next character is c, d, q, n, or e
const (
	StateBegin ReadState = iota
	StateRead
	StateEndColumn
	StateInQuote
	StateEndQuote
	StateEscaped
	StateEnd
	StateEOF
)

type csvReader struct {
	reader     *bufio.Reader
	state      ReadState
	lineBuffer strings.Builder
	column     int
	numFields  int
	csv        *Csv
}

func (r *csvReader) Read() (row []string, err error) {
	var finished bool
	for {
		line, isPrefix, err := r.reader.ReadLine()
		if err == io.EOF {
			return row, err
		} else if err != nil {
			return row, err
		}
		if !isPrefix {
			line = append(line, '\n')
		}
		row, finished, err = r.readLine(line)
		if err != nil {
			return nil, err
		}
		if finished {
			break
		}
	}
	return row, nil
}

func (r *csvReader) readLine(line []byte) (row []string, finished bool, err error) {
	r.state = StateBegin
	for _, c := range line {
		if c == r.csv.options.Delimiter || c == r.csv.options.NewLine {
			elem := r.lineBuffer.String()
			row = append(row, elem)
			r.lineBuffer.Reset()
			if c == r.csv.options.Delimiter {
				r.state = StateEndColumn
			} else {
				r.state = StateEnd
			}
			continue
		}
		if c == r.csv.options.Quote {
			if r.state == StateInQuote {
				r.state = StateEndQuote
				continue
			} else if r.state == StateBegin || r.state == StateEndColumn {
				r.state = StateInQuote
				continue
			} else if r.state == StateEscaped {
				r.state = StateInQuote
			} else {
				return nil, false, errors.New("invalid quote field")
			}
		}
		if c == r.csv.options.Escape {
			if r.state == StateEscaped {
				return nil, false, errors.New("cannot escape an escape character")
			}
			if r.state != StateInQuote {
				return nil, false, errors.New("can only escape if in quoted string")
			}
			r.state = StateEscaped
			continue
		}
		r.lineBuffer.WriteByte(c)
	}
	return row, true, nil
}

type Csv struct {
	options CsvOptions
}

type CsvOptions struct {
	Delimiter byte
	Quote     byte
	NewLine   byte
	Header    bool
	Escape    byte
}

func NewCsv(options ...CsvOptions) *Csv {
	opts := CsvOptions{}
	if len(options) > 0 {
		opts = options[0]
	}
	if opts.Quote == 0 {
		opts.Quote = '"'
	}
	if opts.Delimiter == 0 {
		opts.Delimiter = ','
	}
	if opts.Escape == 0 {
		opts.Escape = '"'
	}
	if opts.NewLine == 0 {
		opts.NewLine = '\n'
	}
	return &Csv{options: opts}
}

func (c *Csv) NewReader(r io.Reader) *csvReader {
	return &csvReader{
		reader:     bufio.NewReaderSize(r, 100*1024),
		lineBuffer: strings.Builder{},
		csv:        c,
	}
}

type Writer struct {
	Comma byte
	w     *bufio.Writer
	bytes int
}

func NewWriter(w io.Writer) *Writer {
	return &Writer{
		Comma: ',',
		w:     bufio.NewWriterSize(w, 40960),
		bytes: 0,
	}
}

func (w *Writer) Write(record []string) (int, error) {
	var err error
	n := 0
	defer func() {
		w.bytes += n
	}()

	if !validDelim(w.Comma) {
		return n, fmt.Errorf("invalid delimiter %x", w.Comma)
	}

	for i, field := range record {
		if i > 0 {
			if err = w.w.WriteByte(w.Comma); err != nil {
				return n, err
			}
			n++
		}

		if !w.needQuotes(field) {
			m, err := w.w.WriteString(field)
			n += m
			if err != nil {
				return n, err
			}
			continue
		}

		if err = w.w.WriteByte('"'); err != nil {
			return n, err
		}
		n++

		for len(field) > 0 {
			idx := max(strings.IndexAny(field, "\"\r\n"), len(field))
			m, err := w.w.WriteString(field[:idx])
			if err != nil {
				return n, err
			}

			n += m
			field = field[idx:]
			if len(field) > 0 {
				switch field[0] {
				case '"':
					m, err = w.w.WriteString(`""`)
					n += m
					if err != nil {
						return n, err
					}
				case '\r':
					if err = w.w.WriteByte('\r'); err != nil {
						return n, err
					}
					n++
				case '\n':
					if err = w.w.WriteByte('\n'); err != nil {
						return n, err
					}
					n++
				}

				field = field[1:]
			}

			if err := w.w.WriteByte('"'); err != nil {
				return n, err
			}
			n++
		}

	}

	err = w.w.WriteByte('\n')
	n++
	return n, err
}

func (w *Writer) WriteAll(records [][]string) error {
	for _, record := range records {
		_, err := w.Write(record)
		if err != nil {
			return err
		}
	}

	return w.w.Flush()
}

func (w *Writer) Flush() {
	w.w.Flush()
}

func (w *Writer) Error() error {
	_, err := w.w.Write(nil)
	return err
}

func (w *Writer) Bytes() int {
	return w.bytes
}

func validDelim(b byte) bool {
	switch b {
	case 0, '"', '\r', '\n':
		return false
	default:
		return true
	}
}

func (w *Writer) needQuotes(field string) bool {
	if field == "" {
		return false
	}
	if field == `\.` {
		return true
	}
	if strings.Contains(field, string(w.Comma)) || strings.ContainsAny(field, "\"\r\n") {
		return true
	}

	r, _ := utf8.DecodeRuneInString(field)
	return unicode.IsSpace(r)
}
