package main

import (
	"CC/algos/algo2adapter"
	"CC/algos/algo_config"
	"CC/flag_handler"
	"CC/graph"
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

	err = adapterFunc(&algoConfig)
	if (err != nil) {
		log.Panicln("main adapterFunc:", err)
	}
}
