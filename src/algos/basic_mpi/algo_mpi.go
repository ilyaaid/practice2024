package basic_mpi

import (
	"CC/src/algos/algo_config"
	"CC/src/graph"

	mpi "github.com/sbromberger/gompi"
)

func Vertex2Proc(conf *algo_config.AlgoConfig, v graph.IndexType) int {
	return int(v% graph.IndexType((conf.ProcNum))) 
}

func getSlaveRanksArray() []int {
	arr := []int{}
	for i := 0; i < MASTER_RANK; i++ {
		arr = append(arr, i)
	}
	return arr
}

func Run(conf *algo_config.AlgoConfig) error {
	rank := mpi.WorldRank()
	// общий коммуникатор
	comm := mpi.NewCommunicator(nil)

	MASTER_RANK = conf.ProcNum
	comm.Barrier()

	// коммуникатор только для ведомых процессов (чтобы ставить барьеры только для них)
	slavesComm := mpi.NewCommunicator(getSlaveRanksArray())

	var master Master
	var slave Slave
	var err error

	if rank == MASTER_RANK {
		master = Master {
			comm: comm,
			conf: conf,
		}
		err = master.Init()
	} else {
		slave = Slave {
			rank: rank,
			comm: comm,
			slavesComm: slavesComm,
			conf: conf,
		}
		err = slave.Init()
	}
	if err != nil {
		return err
	}

	// распеределение ребер с ведущего по всем ведомым узлам (разбиение графа на части)
	if (rank == MASTER_RANK) {
		err = master.SendAllEdges()
	} else {
		err = slave.GetEdges()
	}
	if err != nil {
		return err
	}
	comm.Barrier()

	if (rank != MASTER_RANK) {
		slave.countReceivingPPNumber()
	}

	comm.Barrier()

	// вычисляем CC
	if (rank == MASTER_RANK) {
		err = master.manageCCSearch()
	} else {
		err = slave.CCSearch()
	}
	if err != nil {
		return err
	}

	comm.Barrier()

	// реализация через отправку результата на ведущий процесс
	if (rank == MASTER_RANK) {
		err = master.getResult()
	} else {
		err = slave.sendResult()
	}
	if err != nil {
		return err
	}

	if (rank == MASTER_RANK) {
		err = master.prepResult()
	}
	if err != nil {
		return err
	}
	comm.Barrier()

	if (rank != MASTER_RANK) {
		err = slave.addResult()
	}
	if err != nil {
		return err
	}

	return nil
}
