package algo2run

import (
	"CC/algos/algo_config"
	"CC/algos/algo_types"
	"CC/algos/basic_mpi"
	"CC/algos/fastsv_mpi"
	"fmt"
)

type RunFuncType func(*algo_config.AlgoConfig) error

var algo2run = map[string]RunFuncType {
	algo_types.ALGO_basic_mpi: basic_mpi.Run,
	algo_types.ALGO_fastsv_mpi: fastsv_mpi.Run,
}

func GetRun(algo string) (RunFuncType, error) {
	runFunc, isExist := algo2run[algo]
	if !isExist {
		return nil, fmt.Errorf("algo: \"%s\" does not exist", algo)
	}

	return runFunc, nil
}



