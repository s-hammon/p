package main

import (
	"errors"
	"io"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
	assert.Equal(t, `["movie","director","year"]`, Marshal(rows[0]))
	assert.Equal(t, `["Tombstone","George P. Cosmatos","1993"]`, Marshal(rows[1]))
	assert.Equal(t, `["Yojimbo","Akira Kurosawa","1961"]`, Marshal(rows[2]))
	assert.Equal(t, `["The Thing","John Carpenter","1982"]`, Marshal(rows[3]))

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
