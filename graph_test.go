package p

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestBFS(t *testing.T) {
	graph := NewGraph[string]()
	graph.AddEdge("api-gateway", "auth")
	graph.AddEdge("api-gateway", "orders")
	graph.AddEdge("auth", "users")
	graph.AddEdge("orders", "payments")
	graph.AddEdge("orders", "inventory")

	got := BFS(graph, "api-gateway")
	require.ElementsMatch(t,
		[]string{
			"api-gateway",
			"auth",
			"orders",
			"users",
			"payments",
			"inventory",
		},
		got,
	)

	got = BFS(graph, "orders")
	require.ElementsMatch(t,
		[]string{
			"orders",
			"payments",
			"inventory",
		},
		got,
	)

	got = BFS(graph, "inventory")
	require.ElementsMatch(t,
		[]string{
			"inventory",
		},
		got,
	)
}

func TestDFS(t *testing.T) {
	graph := NewGraph[string]()
	graph.AddEdge("a", "b")
	graph.AddEdge("a", "c")
	graph.AddEdge("c", "d")
	graph.AddEdge("d", "e")

	got := DFS(graph, "c")
	require.ElementsMatch(t,
		[]string{
			"c",
			"d",
			"e",
		},
		got,
	)

	got = DFS(graph, "g")
	require.Len(t, got, 1)
	require.Equal(t, "g", got[0])
}
