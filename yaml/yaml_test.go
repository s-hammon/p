// NOTE: WIP
package yaml

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

var testYaml string = `
world: Ivalice
nations:
  Dalmasca:
    capital: Rabanastre
    climate: desert
  Archadia:
    capital: Archades
    climate: mild
  Nabradia:
    capital: Nabudis
    climate: deadlands
characters:
  pirates:
    - Balthier
    - Fran
    - Reddas
  judges:
    - Gabranth
    - Drace
    - Ghis
    - Bergen
    - Zargabaath
`

var testMap = map[string]any{
	"world": "Ivalice",
	"nations": map[string]any{
		"Dalmasca": map[string]any{
			"capital": "Rabanastre",
			"climate": "desert",
		},
		"Archadia": map[string]any{
			"capital": "Archades",
			"climate": "mild",
		},
		"Nabradia": map[string]any{
			"capital": "Nabudis",
			"climate": "deadlands",
		},
	},
}

func TestDecoder(t *testing.T) {
	dec := NewDecoder(strings.NewReader(testYaml))

	out := make(map[string]any)
	err := dec.Decode(&out)
	require.NoError(t, err)
}
