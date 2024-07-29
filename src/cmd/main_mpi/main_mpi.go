package main

import (
	"CC/algos/algo2run"
	"CC/algos/algo_config"
	"CC/flag_handler"
	"CC/mympi"
	"fmt"
	"log"
	// "os"
)

func main() {
	var err error

	mympi.Start(false)
	defer mympi.Stop()
	
	// Настройка логов
	log.SetPrefix("======= MPI Proc (" + fmt.Sprintf("%d", mympi.WorldRank()) + ") =======\n")
	log.SetFlags(log.Lmsgprefix)

	// TODO попытаться добавить вывод в файл
	// logFile, err := os.Create("logs.txt")
	// if err != nil {
	// 	log.Panicln(err)
	// }
	// defer logFile.Close()

	// log.SetOutput(logFile)


	// Чтение и парсинг флагов
	var fh flag_handler.FlagHadler
	fh.Parse()

	conf, err := algo_config.StrToObj(fh.Conf)
	if err != nil {
		log.Panicln("main_mpi StrToObj:", err)
	}

	runFunc, err := algo2run.GetRun(fh.Algo)
	if err != nil {
		log.Panicln("main_mpi GetRun:", err)
	}

	err = runFunc(conf)
	if err != nil {
		log.Panicln("main_mpi runFunc:", err)
	}
}
