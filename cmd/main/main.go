package main

import (
	"CC/src/algos/algo2adapter"
	"CC/src/algos/algo_config"
	"CC/src/flag_handler"
	"CC/src/graph"
	"log"
	"os"
)


func main() {
	log.SetPrefix("======= main program =======\n")
	log.SetFlags(log.Lmsgprefix)

	// парсинг и вывод флагов командной строки
	var fh flag_handler.FlagHadler
	fh.Parse()
	fh.Print(os.Stdout)

	algo := fh.Algo
	// получаем соответствующую функцию адаптера для получения компонент связности
	adapterFunc, err := algo2adapter.GetAdapter(algo)
	if err != nil {
		log.Panicln(err)
	}

	algoConfig := algo_config.AlgoConfig{
		GrIO: &graph.FileGraphIO{Filename:fh.File},
		ProcNum: fh.Proc,
	}

	graph, err := adapterFunc(&algoConfig)
	if (err != nil) {
		log.Panicln(err)
	}

	log.Println(graph.CC)
}
