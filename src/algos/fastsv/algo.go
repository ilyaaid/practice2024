package fastsv

import (
	"CC/graph"
)

func copyCC(dest map[graph.IndexType]graph.IndexType, src map[graph.IndexType]graph.IndexType) {
	for key, value := range src {
		dest[key] = value
	}
}

func CCSearch(g *graph.Graph) {
	f := g.F

	changed := true

	for changed {
		changed = false

		// stochastic hooking
		for _, edge := range g.Edges {
			u, v := edge.V1, edge.V2

			if f[f[u]] > f[f[v]] {
				f[f[u]] = f[f[v]]
				changed = true
			}
			if f[f[v]] > f[f[u]] {
				f[f[v]] = f[f[u]]
				changed = true
			}
		}
		// aggressive hooking
		for _, edge := range g.Edges {
			u, v := edge.V1, edge.V2

			if f[u] > f[f[v]] {
				f[u] = f[f[v]]
				changed = true
			}

			if f[v] > f[f[u]] {
				f[v] = f[f[u]]
				changed = true
			}
		}
		// shortcut
		for u := range g.F {
			if f[u] > f[f[u]] {
				f[u] = f[f[u]]
				changed = true
			}
		}
	}
}
