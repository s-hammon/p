package p

import "strings"

func IsFileType(fname string, exts ...string) bool {
	for _, ext := range exts {
		if strings.HasSuffix(fname, ext) {
			return true
		}
	}
	return false
}
