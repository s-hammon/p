package p

import (
	"compress/gzip"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
)

func IsFileType(fname string, exts ...string) bool {
	for _, ext := range exts {
		if strings.HasSuffix(fname, ext) {
			return true
		}
	}
	return false
}

func CompressWrite(fname string, r io.Reader) (int, error) {
	if !IsFileType(fname, "gz") {
		return 0, errors.New("file must be .gz")
	}

	f, err := os.Create(fname)
	if err != nil {
		return 0, fmt.Errorf("os.Create: %v", err)
	}

	gw := gzip.NewWriter(f)
	defer func() {
		gw.Close()
		f.Close()
	}()

	n, err := io.Copy(gw, r)
	if err != nil {
		return 0, fmt.Errorf("io.Copy: %v", err)
	}

	return int(n), nil
}
