// NOTE: WIP
package yaml

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParser(t *testing.T) {
	p := newParser()

	p.parse([]string{"version: 2.6"})
	require.Equal(t, stateKey, p.state)

	p.transition("main:")
	require.Equal(t, stateVal, p.state)
	require.Equal(t, 1, p.indentStack.Len())
	top, _ := p.indentStack.Peek()
	require.Equal(t, 0, top)

	p.transition("  name: Sven")
	require.Equal(t, stateScalar, p.state)
	require.Equal(t, 2, p.indentStack.Len())
	top, _ = p.indentStack.Peek()
	require.Equal(t, 2, top)

	p.transition("  job: Developer")
	require.Equal(t, stateNewLine, p.state)
	require.Equal(t, 2, p.indentStack.Len())
	top, _ = p.indentStack.Peek()
	require.Equal(t, 2, top)

	p.transition("secondary:")
	require.Equal(t, stateKey, p.state)
	require.Equal(t, 1, p.indentStack.Len())
	top, _ = p.indentStack.Peek()
	require.Equal(t, 0, top)

	p.transition("  name: King")
	require.Equal(t, stateScalar, p.state)
	require.Equal(t, 2, p.indentStack.Len())
	top, _ = p.indentStack.Peek()
	require.Equal(t, 2, top)

	p.transition("  jobs:")
	require.Equal(t, stateNewLine, p.state)

	p.transition("    - friar")
	require.Equal(t, stateListItem, p.state)
	require.Equal(t, 3, p.indentStack.Len())
	top, _ = p.indentStack.Peek()
	require.Equal(t, 4, top)

	p.transition("   - luchador")
	require.Equal(t, stateErr, p.state)
}
