package basic_mpi

import (
	"CC/algos/algo_config"
	"CC/graph"

	mpi "github.com/sbromberger/gompi"
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

func sendMes(comm *mpi.Communicator, mes []byte, toID int, tag int) {
	comm.SendBytes(mes, toID, tag)
}

func recvMes(comm *mpi.Communicator, fromID int, tag int) ([]byte, mpi.Status){
	mes, status := comm.RecvBytes(fromID, tag)
	for len(mes) > 0 && mes[len(mes) - 1] == 0 {
		mes = mes[:len(mes) - 1]
	}
	return mes, status
}

func sendTag(comm *mpi.Communicator, toID int, tag int) {
	sendMes(comm, []byte{0}, toID, tag)
}

func recvTag(comm *mpi.Communicator, fromID int, tag int) mpi.Status {
	_, status := recvMes(comm, fromID, tag)
	return status
}
