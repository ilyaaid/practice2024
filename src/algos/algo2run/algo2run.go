package algo2run

import (
	"CC/src/algos/algo_config"
	"CC/src/algos/algo_types"
	"CC/src/algos/basic_mpi"
	"log"
)

type RunFuncType func(*algo_config.AlgoConfig)

var algo2run = map[string]RunFuncType {
	algo_types.ALGO_basic_mpi: basic_mpi.Run,
}

func GetRun(algo string) RunFuncType {
	runFunc, isExist := algo2run[algo]
	if !isExist {
		log.Panic("algo: \"" + algo + "\" does not exist")
	}

	return runFunc
}



