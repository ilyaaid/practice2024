package basic

import (
	"CC/src/algos/algo_config"
	"CC/src/graph"
)

func _CCSearch(g *graph.Graph) {
	f := make(map[graph.IndexType]graph.IndexType)

	for i := graph.IndexType(0); i < g.CntVertex; i++ {
		f[i] = i
	}

	changed := true
	for changed {
		changed = false

		for _, edge := range g.Edges {
			if (f[edge.V1] < f[edge.V2]) {
				f[edge.V2] = f[edge.V1]
				changed = true
			} else if (edge.V2 < f[edge.V1]) {
				f[edge.V1] = f[edge.V2]
				changed = true
			}
		}
	}

	g.CC = f
}

func Adapter(conf *algo_config.AlgoConfig) *graph.Graph {
	g := graph.ReadFromFile(conf.GraphFilename)

	_CCSearch(g)
	return g
}
