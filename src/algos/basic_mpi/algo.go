package basic_mpi

import (
	"CC/algos/algo_config"
	"CC/algos/ilogger"
	"CC/graph"
	"CC/mympi"
	"CC/utils"
	"fmt"
	"log"
)

// Варианты алгоритма
const (
	// непрерывное обновление
	VARIANT_CONT = "cont"
)

var variants []string = []string{
	// VARIANT_DISC,
	VARIANT_CONT,
}

type Algo struct {
	logger *Logger
	conf   *algo_config.AlgoConfig
}

func (algo *Algo) GetLogger() ilogger.ILogger {
	return algo.logger
}

func (algo *Algo) Init(conf *algo_config.AlgoConfig) error {
	algo.conf = conf
	MASTER_RANK = algo.conf.ProcNum

	algo.logger = &Logger{}
	algo.logger.Init()

	// проверка верности варианта алгоритма
	if !utils.Contains(variants, algo.conf.Variant) {
		return fmt.Errorf("unknown algo variant=%s", algo.conf.Variant)
	}
	return nil
}

func (algo *Algo) Close() error {
	algo.logger.Close()
	return nil
}

func (algo *Algo) getSlave(v graph.IndexType) int {
	return int(v % graph.IndexType((algo.conf.ProcNum)))
}

func (algo *Algo) Run() error {
	rank := mympi.WorldRank()
	// общий коммуникатор
	comm := mympi.WorldCommunicator()

	// коммуникатор только для ведомых процессов (чтобы ставить барьеры только для них)
	slavesComm := mympi.SlavesCommunicator(MASTER_RANK)

	var master Master
	var slave Slave
	var err error

	if rank == MASTER_RANK {
		master = Master{
			comm: comm,
			algo: algo,
		}
		err = master.Init()
	} else {
		slave = Slave{
			rank:       rank,
			comm:       comm,
			slavesComm: slavesComm,
			algo:       algo,
		}
		err = slave.Init()
	}
	if err != nil {
		return err
	}

	// распеределение ребер с ведущего по всем ведомым узлам (разбиение графа на части)
	if rank == MASTER_RANK {
		err = master.SendAllEdges()
	} else {
		err = slave.GetEdges()
	}
	if err != nil {
		return err
	}

	comm.Barrier()

	// подсчет количества получаемых сообщений на обновление родителя
	if rank != MASTER_RANK {
		slave.countReceivingPPNumber()
	}

	comm.Barrier()

	if rank == MASTER_RANK {
		log.Println("==================")
	}

	// вычисляем CC
	if rank == MASTER_RANK {
		err = master.manageCCSearch()
		if err != nil {
			log.Panicln("master.manageCCSearch:", err)
			return err
		}
	} else {
		err = slave.CCSearch()
		if err != nil {
			log.Panicln("slave.CCSearch:", err)
			return err
		}
	}

	comm.Barrier()

	// реализация через отправку результата на ведущий процесс
	if rank == MASTER_RANK {
		err = master.getResult()
	} else {
		err = slave.sendResult()
	}
	if err != nil {
		return err
	}

	comm.Barrier()

	if rank == MASTER_RANK {
		log.Println(master.g.F)
	}
	return nil
}
