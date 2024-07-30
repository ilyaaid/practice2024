package basic_mpi

import (
	"CC/algos/algo_config"
	"CC/graph"
	"CC/mympi"
	"log"
)

func Vertex2Proc(conf *algo_config.AlgoConfig, v graph.IndexType) int {
	return int(v % graph.IndexType((conf.ProcNum)))
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

	// if rank != MASTER_RANK {
	// 	log.Println(len(slave.edges))
	// } else {
	// 	log.Println(len(master.g.Edges))
	// }

	comm.Barrier()

	// подсчет количества получаемых сообщений на обновление родителя
	if (rank != MASTER_RANK) {
		slave.countReceivingPPNumber()
	}

	comm.Barrier()

	//тест
	// if rank == MASTER_RANK {
	// 	log.Println("sending")
	// 	sendMes(master.comm, []byte{234, 24, 54, 56}, 0, TAG_IS_CHANGED)
	// 	log.Println("sended")
	// } else {
	// 	if (slave.rank == 0) {
	// 		is := false
	// 		var status *mympi.Status
	// 		for !is {
	// 			is, status = slave.comm.Iprobe(MASTER_RANK, TAG_IS_CHANGED)
	// 		}
	// 		log.Println(is, status.GetCount(mympi.Byte))
	// 		status = slave.comm.Probe(MASTER_RANK, TAG_IS_CHANGED)
	// 		log.Println(status.GetCount(mympi.Byte))
	// 		mes, _ := slave.comm.RecvBytes(MASTER_RANK, TAG_IS_CHANGED)
	// 		log.Println(mes)
	// 		// buf := make([]byte, 1)
	// 		// slave.comm.RecvPreallocBytes(buf, MASTER_RANK, TAG_IS_CHANGED)
	// 		// log.Println(buf)
	// 		// recvMes(slave.comm, MASTER_RANK, TAG_IS_CHANGED)
	// 	}
	// }

	if rank == MASTER_RANK {
		log.Println("==================")
	}

	// вычисляем CC
	if (rank == MASTER_RANK) {
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

	// log.Println(mympi.X_min, mympi.Len_min, mympi.Len_max)

	comm.Barrier()

	// if rank != MASTER_RANK {
	// 	log.Println(slave.cc)
	// }

	// реализация через отправку результата на ведущий процесс
	if (rank == MASTER_RANK) {
		err = master.getResult()
	} else {
		err = slave.sendResult()
	}
	if err != nil {
		return err
	}

	comm.Barrier()

	// if (rank == MASTER_RANK) {
	// 	err = master.prepResult()
	// }
	// if err != nil {
	// 	return err
	// }

	// comm.Barrier()

	// if (rank != MASTER_RANK) {
	// 	err = slave.addResult()
	// }
	// if err != nil {
	// 	return err
	// }

	if (rank == MASTER_RANK) {
		log.Println(master.g.CC)
	}
	return nil
}
