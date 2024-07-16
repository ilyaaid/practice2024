package basic

import (
	"CC/src/algos/algo_config"
	"CC/src/graph"
)

func Adapter(conf *algo_config.AlgoConfig) *graph.Graph {
	g := graph.ReadFromFile(conf.GraphFilename)

	CCSearch(g)
	return g
}
