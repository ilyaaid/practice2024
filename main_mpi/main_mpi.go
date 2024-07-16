package main

import (
	"CC/src/algos/algo2run"
	"CC/src/algos/algo_config"
	"CC/src/flag_handler"
	"fmt"
	"log"

	"github.com/sbromberger/gompi"
)

func main() {
	mpi.Start(false)
	defer mpi.Stop()

	// Настройка логов
	log.SetPrefix("======= MPI Proc (" + fmt.Sprintf("%d", mpi.WorldRank()) + ") =======\n")
	log.SetFlags(log.Lmsgprefix)

	// Чтение и парсинг флагов
	var fh flag_handler.FlagHadler
	fh.Parse()

	conf := algo_config.StrToObj(fh.Conf)
	algo2run.GetRun(fh.Algo)(conf)
}
