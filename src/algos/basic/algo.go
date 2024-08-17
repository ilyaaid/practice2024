package basic

import (
	"CC/graph"
)

func CCSearch(g *graph.Graph) {
	f := g.F

	changed := true
	for changed {
		changed = false

		for _, edge := range g.Edges {
			if f[edge.V1] < f[edge.V2] {
				f[edge.V2] = f[edge.V1]
				changed = true
			} else if f[edge.V2] < f[edge.V1] {
				f[edge.V1] = f[edge.V2]
				changed = true
			}
		}
	}
}
