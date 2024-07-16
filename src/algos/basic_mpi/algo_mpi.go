package basic_mpi

import (
	"CC/src/algos/algo_config"
	"CC/src/graph"
	"log"

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

func Run(conf *algo_config.AlgoConfig) {
	rank := mpi.WorldRank()
	// общий коммуникатор
	comm := mpi.NewCommunicator(nil)

	MASTER_RANK = conf.ProcNum
	comm.Barrier()

	// коммуникатор только для ведомых процессов (чтобы ставить барьеры только для них)
	slavesComm := mpi.NewCommunicator(getSlaveRanksArray())

	var master Master
	var slave Slave

	if rank == MASTER_RANK {
		master = Master {
			comm: comm,
			conf: conf,
		}
		master.Init()
	} else {
		slave = Slave {
			rank: rank,
			comm: comm,
			slavesComm: slavesComm,
			conf: conf,
		}
		slave.Init()
	}

	// распеределение ребер с ведущего по всем ведомым узлам (разбиение графа на части)
	if (rank == MASTER_RANK) {
		master.SendAllEdges()
	} else {
		slave.GetEdges()
		log.Println(slave.edges)
	}
	comm.Barrier()

	if (rank != MASTER_RANK) {
		slave.countReceivingPPNumber()
	}

	comm.Barrier()

	// вычисляем CC
	if (rank == MASTER_RANK) {
		master.manageCCSearch()
	} else {
		slave.CCSearch()
	}

	comm.Barrier()
	if (rank == MASTER_RANK) {
		master.getResult()
	} else {
		slave.sendResult()
	}
	// if rank != MASTER_RANK {
	// 	log.Println(slave.cc)
	// }
	
}
