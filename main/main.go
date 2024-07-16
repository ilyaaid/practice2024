package main

import (
	"CC/src/algos/algo2adapter"
	"CC/src/algos/algo_config"
	"CC/src/flag_handler"
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
	adapterFunc := algo2adapter.GetAdapter(algo)

	algoConfig := algo_config.AlgoConfig{
		GraphFilename: fh.File,
		ProcNum: fh.Proc,
	}

	graph := adapterFunc(&algoConfig)
	log.Println(graph.CC)

}
