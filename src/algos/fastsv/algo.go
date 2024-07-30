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
	f := g.CC
	fnext := make(map[graph.IndexType]graph.IndexType) 
	copyCC(fnext, f)

	changed := true

	for changed {
		changed = false

		// stoch
		for _, edge := range g.Edges {
			u, v := edge.V1, edge.V2

			if fnext[f[u]] > f[f[v]] {
				fnext[f[u]] = f[f[v]]
				changed = true
			}
			if fnext[f[v]] > f[f[u]] {
				fnext[f[v]] = f[f[u]]
				changed = true 
			}
		}

		// aggr 
		for _, edge := range g.Edges {
			u, v := edge.V1, edge.V2

			if fnext[u] > f[f[v]] {
				fnext[u] = f[f[v]]
				changed = true
			}

			if fnext[v] > f[f[u]] {
				fnext[v] = f[f[u]]
				changed = true
			}
		}

		// shortcut
		for u := range g.CC {
			if fnext[u] > f[f[u]] {
				fnext[u] = f[f[u]]
				changed = true
			}
		}
		
		copyCC(f, fnext)
	}
}
