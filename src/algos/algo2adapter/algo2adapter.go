package algo2adapter

import (
	"CC/src/algos/algo_config"
	"CC/src/algos/algo_types"
	"CC/src/algos/basic"
	"CC/src/algos/basic_mpi"
	"CC/src/graph"
	"fmt"
)

type AdapterFuncType func(*algo_config.AlgoConfig) (*graph.Graph, error)

var algo2adapter = map[string]AdapterFuncType{
	algo_types.ALGO_basic:     basic.Adapter,
	algo_types.ALGO_basic_mpi: basic_mpi.Adapter,
}

func GetAdapter(algo string) (AdapterFuncType, error) {
	adapterFunc, isExistAdapter := algo2adapter[algo]
	if !isExistAdapter {
		return nil, fmt.Errorf("algo: \"%s\" does not exist", algo)
	}

	return adapterFunc, nil
}
