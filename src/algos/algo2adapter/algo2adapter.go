package algo2adapter

import (
	"CC/algos/algo_config"
	"CC/algos/algo_types"
	"CC/algos/basic"
	"CC/algos/basic_mpi"
	"CC/algos/fastsv_mpi"
	"fmt"
)

type AdapterFuncType func(*algo_config.AlgoConfig) (error)

var algo2adapter = map[string]AdapterFuncType{
	algo_types.ALGO_basic:     basic.Adapter,
	algo_types.ALGO_basic_mpi: basic_mpi.Adapter,
	algo_types.ALGO_fastsv_mpi: fastsv_mpi.Adapter,
	
}

func GetAdapter(algo string) (AdapterFuncType, error) {
	adapterFunc, isExistAdapter := algo2adapter[algo]
	if !isExistAdapter {
		return nil, fmt.Errorf("algo: \"%s\" does not exist", algo)
	}

	return adapterFunc, nil
}
