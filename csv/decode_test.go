package csv

import (
	"bytes"
	"testing"
	"time"

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

func TestDecodeLine_TimeFormats(t *testing.T) {
	tests := []struct {
		name string
		csv  string
		want string
	}{
		{
			name: "RFC3339",
			csv:  "date\n2005-01-03T14:22:00Z\n",
			want: "2005-01-03T14:22:00Z",
		},
		{
			name: "YYYY-MM-DD",
			csv:  "date\n2005-01-03\n",
			want: "2005-01-03T00:00:00Z",
		},
		{
			name: "YYYYMMDDhhmmss",
			csv:  "date\n20050103142200\n",
			want: "2005-01-03T14:22:00Z",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf := bytes.NewBufferString(tt.csv)

			dec, err := NewDecoder(buf)
			require.NoError(t, err)

			type record struct {
				Date time.Time `csv:"date"`
			}

			var rec record
			require.NoError(t, dec.Decode(&rec))
			require.Equal(t, tt.want, rec.Date.UTC().Format(time.RFC3339))
		})
	}
}
