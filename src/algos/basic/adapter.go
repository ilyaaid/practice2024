package basic

import (
	"CC/src/algos/algo_config"
	"CC/src/graph"
)

func Adapter(conf *algo_config.AlgoConfig) (*graph.Graph, error) {
	g, err := conf.GrIO.Read()
	if err != nil {
		return nil, err
	}
	CCSearch(g)
	return g, nil
}
