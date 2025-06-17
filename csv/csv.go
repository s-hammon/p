package csv

import (
	"bufio"
	"errors"
	"io"
	"strings"
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

func (c *Csv) NewWriter(w io.Writer) *Writer {
	return &Writer{
		Comma: c.options.Delimiter,
		w:     bufio.NewWriterSize(w, 100*1024),
		bytes: 0,
	}
}
