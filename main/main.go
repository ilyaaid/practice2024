package main

import (
	"CC/src/algos/algo_adapter"
	"CC/src/algos/algo_config"
	"CC/src/flag_handler"
	"log"
	"os"
)

func main() {
	// парсинг и вывод флагов командной строки
	var fh flag_handler.FlagHadler
	fh.Parse()
	fh.Print(os.Stdout)

	algo := fh.Algo
	// получаем соответствующую функцию адаптера для получения компонент связности
	adapterFunc := algo_adapter.GetAdapter(algo)

	algoConfig := algo_config.AlgoConfig{
		GraphFilename: fh.File,
		Proc: fh.Proc,
	}

	graph := adapterFunc(&algoConfig)
	log.Println(graph.CC)

}
