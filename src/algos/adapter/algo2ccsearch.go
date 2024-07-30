package adapter

import (
	"CC/algos/algo_types"
	"CC/algos/basic"
	"CC/algos/fastsv"
	"CC/graph"
	"fmt"
)

type CCSearchFuncType func(*graph.Graph)

var algo2ccsearch = map[string]CCSearchFuncType{
	algo_types.ALGO_basic:  basic.CCSearch,
	algo_types.ALGO_fastsv: fastsv.CCSearch,
}

func GetCCSearchFunc(algo string) (CCSearchFuncType, error) {
	function, isExist := algo2ccsearch[algo]
	if !isExist {
		return nil, fmt.Errorf("algo: \"%s\" does not exist", algo)
	}

	return function, nil
}
