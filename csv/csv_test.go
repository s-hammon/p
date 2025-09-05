package csv

import (
	"bytes"
	"errors"
	"io"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/s-hammon/p"
)

func TestCsvReader(t *testing.T) {
	data := `movie,director,year
"Tombstone","George P. Cosmatos",1993
Yojimbo,"Akira Kurosawa",1961
"The Thing","John Carpenter",1982`

	csv := NewCsv()
	r := csv.NewReader(strings.NewReader(data))

	rows := [][]string{}
	for {
		row, err := r.Read()
		if err == io.EOF {
			break
		}
		require.NoError(t, err)
		rows = append(rows, row)
	}

	require.Equal(t, 4, len(rows))
	assert.Equal(t, `["movie","director","year"]`, p.Marshal(rows[0]))
	assert.Equal(t, `["Tombstone","George P. Cosmatos","1993"]`, p.Marshal(rows[1]))
	assert.Equal(t, `["Yojimbo","Akira Kurosawa","1961"]`, p.Marshal(rows[2]))
	assert.Equal(t, `["The Thing","John Carpenter","1982"]`, p.Marshal(rows[3]))

	data = `title|composer|year
"Yell \"Dead Cell\""|Norihiko Hibino|2001
"Into the Mirror"|Minus the Bear|2010`

	csv = NewCsv(CsvOptions{Delimiter: '|', Escape: '\\'})
	r = csv.NewReader(strings.NewReader(data))

	rows = [][]string{}
	for {
		row, err := r.Read()
		if err == io.EOF {
			break
		}
		require.NoError(t, err)
		rows = append(rows, row)
	}

	require.Equal(t, 3, len(rows))
}

func benchmarkReadNew(b *testing.B, opts CsvOptions, rows string) {
	b.ReportAllocs()
	r := NewCsv(opts).NewReader(&nTimes{s: rows, n: b.N})

	for {
		_, err := r.Read()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkReadNew(b *testing.B) {
	benchmarkReadNew(b, CsvOptions{}, benchmarkCSVData)
}

// Copyright 2011 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

const benchmarkCSVData = `x,y,z,w
x,y,z,
x,y,,
x,,,
,,,
"x","y","z","w"
"x","y","z",""
"x","y","",""
"x","","",""
"","","",""
`

type nTimes struct {
	s   string
	n   int
	off int
}

func (r *nTimes) Read(p []byte) (n int, err error) {
	for {
		if r.n <= 0 || r.s == "" {
			return n, io.EOF
		}
		n0 := copy(p, r.s[r.off:])
		p = p[n0:]
		n += n0
		r.off += n0
		if r.off == len(r.s) {
			r.off = 0
			r.n--
		}
		if len(p) == 0 {
			return
		}
	}
}

func TestWrite(t *testing.T) {
	tests := []struct {
		name   string
		input  [][]string
		output string
		err    error
		comma  byte
	}{
		{
			name:   "one blank",
			input:  [][]string{{""}},
			output: "\n",
		},
		{
			name:   "one record two blank",
			input:  [][]string{{"", ""}},
			output: ",\n",
		},
		{
			name:   "two record two blank",
			input:  [][]string{{"", ""}, {"", ""}},
			output: ",\n,\n",
		},
		{
			name:   "two record mixed",
			input:  [][]string{{"a", ""}, {"1", ""}},
			output: "a,\n1,\n",
		},
		{
			name:   "one line",
			input:  [][]string{{"abc"}},
			output: "abc\n",
		},
		{
			name:   "record w/ delim",
			input:  [][]string{{"abc,123"}},
			output: "\"abc,123\"\n",
		},
		{
			name:   "record w/ quote",
			input:  [][]string{{`a"b"`}},
			output: "\"a\"b\"\"\n",
		},
		{
			name:   "record w/ quote 2",
			input:  [][]string{{`"abc"`}},
			output: "\"\"abc\"\"\n",
		},
		{
			name:   "record w/ quote 3",
			input:  [][]string{{`a"b`}},
			output: "\"a\"b\"\n",
		},
		{
			name:   "record w/ newline",
			input:  [][]string{{"abc\ndef"}},
			output: "\"abc\ndef\"\n",
		},
		{
			name:   "two lines",
			input:  [][]string{{"abc", "123"}},
			output: "abc,123\n",
		},
		{
			name:   "two records",
			input:  [][]string{{"abc"}, {"123"}},
			output: "abc\n123\n",
		},
	}

	for _, tt := range tests {
		b := bytes.NewBuffer(nil)
		t.Run(tt.name, func(t *testing.T) {
			f := NewWriter(b)
			f.Comma = p.Coalesce(tt.comma, ',')
			err := f.WriteAll(tt.input)
			if tt.err != nil {
				require.Error(t, err)
				require.True(t, errors.Is(err, tt.err))
			} else {
				require.NoError(t, err)
				require.Equal(t, tt.output, b.String())
			}
		})
	}
}
