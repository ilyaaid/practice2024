package fastsv_mpi

import (
	"CC/algos/algo_config"
	"CC/graph"
	"CC/mympi"
	"log"
	// "log"
)

func Vertex2Proc(conf *algo_config.AlgoConfig, v graph.IndexType) int {
	return int(v % graph.IndexType((conf.ProcNum)))
}

func getSlaveRanksArray() []int {
	arr := []int{}
	for i := 0; i < MASTER_RANK; i++ {
		arr = append(arr, i)
	}
	return arr
}

func copyCC(dest map[graph.IndexType]graph.IndexType, src map[graph.IndexType]graph.IndexType) {
	for key, value := range src {
		dest[key] = value
	}
}

func Run(conf *algo_config.AlgoConfig) error {
	rank := mympi.WorldRank()
	// общий коммуникатор
	comm := mympi.WorldCommunicator()

	MASTER_RANK = conf.ProcNum
	comm.Barrier()

	// коммуникатор только для ведомых процессов (чтобы ставить барьеры только для них)
	slavesComm := mympi.SlavesCommunicator(MASTER_RANK)

	var master Master
	var slave Slave
	var err error

	if rank == MASTER_RANK {
		master = Master{
			comm: comm,
			conf: conf,
		}
		err = master.Init()
	} else {
		slave = Slave{
			rank:       rank,
			comm:       comm,
			slavesComm: slavesComm,
			conf:       conf,
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

	// if rank != MASTER_RANK {
	// 	log.Println(slave.cc)
	// }

	return nil
}
