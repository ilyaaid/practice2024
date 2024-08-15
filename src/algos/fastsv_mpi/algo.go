package fastsv_mpi

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
	// дискретное обновление с учетом родителей первого порядка
	VARIANT_DISC_F_PAR = "disc_f_par"
	// непрерывное обновление с учетом родителей первого порядка
	VARIANT_CONT_F_PAR = "cont_f_par"
	// непрерывное обновление с учетом родителей первого и второго порядка
	VARIANT_CONT_S_PAR = "cont_s_par"
)

var variants []string = []string{
	VARIANT_DISC_F_PAR,
	VARIANT_CONT_F_PAR,
	VARIANT_CONT_S_PAR,
} 

type Algo struct {
	logger *Logger
	conf *algo_config.AlgoConfig
}

func (algo *Algo) Init(conf *algo_config.AlgoConfig) error {
	// сохранение конфигурации
	algo.conf = conf
	MASTER_RANK = algo.conf.ProcNum

	// инициализация логгера
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

func (algo *Algo) GetLogger() ilogger.ILogger {
	return algo.logger
}

func (algo *Algo) Run() error {
	rank := mympi.WorldRank()
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

	if rank == MASTER_RANK {
		log.Println("distributed edges")
	} else {
		// log.Println(slave.edges)
	}

	comm.Barrier()

	if rank == MASTER_RANK {
		err = master.manageCCSearch()
	} else {
		err = slave.runSteps()
	}
	if err != nil {
		return err
	}

	comm.Barrier()

	if rank == MASTER_RANK {
		master.getResult()
	} else {
		slave.sendResult()
	}

	comm.Barrier()

	return nil
}

func (algo *Algo) getSlave(v graph.IndexType) int {
	return int(v % graph.IndexType((algo.conf.ProcNum)))
}

func copyCC(dest map[graph.IndexType]graph.IndexType, src map[graph.IndexType]graph.IndexType) {
	for key, value := range src {
		dest[key] = value
	}
}
