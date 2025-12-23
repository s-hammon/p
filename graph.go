package p

type Graph[T comparable] struct {
	nodes map[T][]T
}

func BFS[T comparable](g *Graph[T], start T) []T {
	visited := make(map[T]bool)
	queue := []T{start}
	res := []T{}

	visited[start] = true
	for len(queue) > 0 {
		node := queue[0]
		queue = queue[1:]

		res = append(res, node)
		for _, neighbor := range g.nodes[node] {
			if !visited[neighbor] {
				visited[neighbor] = true
				queue = append(queue, neighbor)
			}
		}
	}

	return res
}
