package algo2adapter

import (
	"CC/src/algos/algo_config"
	"CC/src/algos/algo_types"
	"CC/src/algos/basic"
	"CC/src/algos/basic_mpi"
	"CC/src/graph"
	"log"
)

type AdapterFuncType func(*algo_config.AlgoConfig) *graph.Graph

var algo2adapter = map[string]AdapterFuncType{
	algo_types.ALGO_basic:     basic.Adapter,
	algo_types.ALGO_basic_mpi: basic_mpi.Adapter,
}

func GetAdapter(algo string) AdapterFuncType {
	adapterFunc, isExistAdapter := algo2adapter[algo]
	if !isExistAdapter {
		log.Panic("algo: \"" + algo + "\" does not exist")
	}

	return adapterFunc
}
