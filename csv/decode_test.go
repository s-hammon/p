package csv

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDecodeLine(t *testing.T) {
	buf := bytes.NewBuffer([]byte("a,b\n12,\"[34,56]\""))
	dec, err := NewDecoder(buf)
	require.NoError(t, err)
	require.Equal(t, map[string]int{"a": 0, "b": 1}, dec.headers)

	type record struct {
		First  string   `csv:"a"`
		Second []string `csv:"b"`
	}
	rec := record{}
	err = dec.Decode(&rec)
	require.NoError(t, err)
	require.Equal(t, record{"12", []string{"34", "56"}}, rec)
}
