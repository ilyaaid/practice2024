package algo_mpi_manager

import (
	"CC/algos/algo_config"

	mpi "github.com/sbromberger/gompi"
)

type AlgoMPIManager struct {
	Rank int
	Comm *mpi.Communicator
	SlavesComm *mpi.Communicator

	Conf *algo_config.AlgoConfig

}
