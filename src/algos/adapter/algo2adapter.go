package adapter

import (
	"CC/algos/algo_config"
	"CC/algos/algo_types"
	"fmt"
)

type AdapterFuncType func(string, *algo_config.AlgoConfig) (error)

var algo2adapter = map[string]AdapterFuncType{
	algo_types.ALGO_basic:     Adapter,
	algo_types.ALGO_basic_mpi: AdapterMPI,
	algo_types.ALGO_fastsv: Adapter,
	algo_types.ALGO_fastsv_mpi: AdapterMPI,
}

func GetAdapter(algo string) (AdapterFuncType, error) {
	adapterFunc, isExistAdapter := algo2adapter[algo]
	if !isExistAdapter {
		return nil, fmt.Errorf("algo: \"%s\" does not exist", algo)
	}

	return adapterFunc, nil
}
