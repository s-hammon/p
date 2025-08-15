package p

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestStack(t *testing.T) {
	stack := NewStack[int]()
	require.NotNil(t, stack)
	require.Empty(t, stack.elements)

	stack.Push(1)
	require.NotEmpty(t, stack.elements)
	require.Len(t, stack.elements, 1)

	stack.Push(2)
	stack.Push(3)
	stack.Push(4)
	stack.Push(5)
	require.Len(t, stack.elements, 5)
	require.Equal(t, []int{1, 2, 3, 4, 5}, stack.elements)

	num, ok := stack.Pop()
	require.True(t, ok)
	require.Equal(t, 5, num)
	require.Len(t, stack.elements, 4)
	require.Equal(t, 4, stack.elements[3])

	for len(stack.elements) > 0 {
		stack.Pop()
	}
	require.Len(t, stack.elements, 0)
}
