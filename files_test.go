package p

import (
	"compress/gzip"
	"io"
	"net/http"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCompressWrite_TheWhale(t *testing.T) {
	const (
		url  = "https://www.gutenberg.org/files/2701/2701-0.txt"
		file = "moby_dick.gz"
	)

	t.Cleanup(func() {
		_ = os.Remove(file)
	})

	resp, err := http.Get(url)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	defer resp.Body.Close()

	n, err := CompressWrite(file, resp.Body)
	require.NoError(t, err)
	require.Greater(t, n, 0)
	t.Log("bytes written:", n)

	info, err := os.Stat(file)
	require.NoError(t, err)
	require.Greater(t, info.Size(), int64(0))

	f, err := os.Open(file)
	require.NoError(t, err)
	defer f.Close()

	gr, err := gzip.NewReader(f)
	require.NoError(t, err)
	defer gr.Close()

	data, err := io.ReadAll(gr)
	require.NoError(t, err)
	require.Greater(t, len(data), 0)
}
