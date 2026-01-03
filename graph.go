package p

type Graph[T comparable] struct {
	nodes map[T]map[T]struct{}
}

func NewGraph[T comparable]() *Graph[T] {
	return &Graph[T]{nodes: make(map[T]map[T]struct{})}
}

func (g Graph[T]) AddNode(node T) {
	if _, ok := g.nodes[node]; !ok {
		g.nodes[node] = make(map[T]struct{})
	}
}

// Adds a directed edge from the first node to the second node
func (g Graph[T]) AddEdge(from, to T) {
	g.AddNode(from)
	g.AddNode(to)

	g.nodes[from][to] = struct{}{}
}

func (g Graph[T]) AddUndirectedEdge(a, b T) {
	g.AddEdge(a, b)
	g.AddEdge(b, a)
}

// Breadth-first search
func BFS[T comparable](g *Graph[T], start T) []T {
	visited := make(map[T]bool)
	queue := []T{start}
	res := []T{}

	visited[start] = true
	for len(queue) > 0 {
		node := queue[0]
		queue = queue[1:]

		res = append(res, node)
		for neighbor := range g.nodes[node] {
			if !visited[neighbor] {
				visited[neighbor] = true
				queue = append(queue, neighbor)
			}
		}
	}

	return res
}

// Depth-first search
func DFS[T comparable](g *Graph[T], start T) []T {
	visited := make(map[T]bool)
	stack := []T{start}
	res := []T{}
	for len(stack) > 0 {
		node := stack[len(stack)-1]
		stack = stack[:len(stack)-1]

		if visited[node] {
			continue
		}

		visited[node] = true
		res = append(res, node)

		for neighbor := range g.nodes[node] {
			if !visited[neighbor] {
				stack = append(stack, neighbor)
			}
		}
	}

	return res
}
