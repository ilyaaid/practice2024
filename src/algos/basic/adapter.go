package basic

import (
	"CC/algos/algo_config"
	"CC/graph"
)

func Adapter(conf *algo_config.AlgoConfig) (*graph.Graph, error) {
	g, err := conf.GrIO.Read()
	if err != nil {
		return nil, err
	}
	CCSearch(g)
	return g, nil
}
