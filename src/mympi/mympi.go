package mympi

/*
#include "mpi/mpi.h"

#cgo linux LDFLAGS: -pthread -L/usr/lib/x86_64-linux-gnu/openmpi/lib -lmpi
*/
import "C"
import "unsafe"

func Start() {
	C.MPI_Init(nil, nil)
}

func Stop() {
	C.MPI_Finalize()
}

func WorldRank() (rank int) {
	var r int32
	C.MPI_Comm_rank(C.MPI_COMM_WORLD, (*C.int)(unsafe.Pointer(&r)))
	return int(r)
}

type Communicator struct {
	comm   C.MPI_Comm
}

func WorldCommunicator() *Communicator {
	var o Communicator

	o.comm = C.MPI_COMM_WORLD
	return &o
}

func SlavesCommunicator(master_rank int) *Communicator {
	var o Communicator

	rank := WorldRank()
	var color C.int
	if rank == master_rank {
		color = C.MPI_UNDEFINED
	} else {
		color = 1
	}
	
	C.MPI_Comm_split(
		C.MPI_COMM_WORLD,
		color,
		(C.int)(rank),
		&o.comm,
	)

	return &o
}

func (o *Communicator) Barrier() {
	C.MPI_Barrier(o.comm)
}
