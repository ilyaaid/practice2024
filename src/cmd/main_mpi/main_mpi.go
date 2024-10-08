package main

import (
	"CC/algos/algo_config"
	"CC/algos/algo_types"
	"CC/algos/basic_mpi"
	"CC/algos/fastsv_mpi"
	"CC/algos/ialgo"
	"CC/flag_handler"
	"CC/mympi"
	"fmt"
	"log"
)

func getAlgo(algo string) ialgo.IAlgo {
	switch algo {
	case algo_types.ALGO_fastsv_mpi:
		return &fastsv_mpi.Algo{}
	case algo_types.ALGO_basic_mpi:
		return &basic_mpi.Algo{}
	default:
		log.Panicln("unkonown algo with MPI")
	}

	return nil
}

func main() {
	mympi.Start(false)
	defer mympi.Stop()

	var err error

	// Настройка логов
	log.SetPrefix("======= MPI Proc (" + fmt.Sprintf("%d", mympi.WorldRank()) + ") =======\n")
	log.SetFlags(log.Lmsgprefix)

	// Чтение и парсинг флагов
	var fh flag_handler.FlagHadler
	fh.Parse()

	conf, err := algo_config.StrToObj(fh.Conf)
	if err != nil {
		log.Panicln("main_mpi StrToObj:", err)
	}

	// fh - объект для хранения значения всех флагов командной строки
	var algo ialgo.IAlgo = getAlgo(fh.Algo)
	err = algo.Init(conf)
	if err != nil {
		log.Panicln("algo Init:", err)
	}

	defer algo.Close()

	if conf.Logging {
		logger := algo.GetLogger()
		logger.Start()
		err = algo.Run()
		logger.Finish()
	} else {
		err = algo.Run()
	}

	if err != nil {
		log.Panicln("main_mpi runFunc:", err)
	}

}
