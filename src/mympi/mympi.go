package mympi

/*
#include <mpi.h>

#cgo linux LDFLAGS: -pthread -L/usr/local/openmpi/lib -lmpi
*/
import "C"
import (
	"log"
	"math"
	"unsafe"
)

const (
	AnySource = C.MPI_ANY_SOURCE
	AnyTag    = C.MPI_ANY_TAG
)

type DataType uint8

const (
	// These constants represent (a subset of) MPI datatypes.
	Byte    DataType = iota
	Uint             // This maps to a uint32 in go.
	Int              // This maps to an int32 in go.
	Ulong            // This maps to a uint64 in go.
	Long             // This maps to an int64 in go.
	Float            // This maps to a float32 in go
	Double           // This maps to a float64 in go.
	Complex          // This maps to a complex128 in go.
)

var dataTypes = [...]C.MPI_Datatype{
	C.MPI_BYTE,
	C.MPI_UINT32_T,
	C.MPI_INT32_T,
	C.MPI_UINT64_T,
	C.MPI_INT64_T,
	C.MPI_FLOAT,
	C.MPI_DOUBLE,
	C.MPI_DOUBLE_COMPLEX,
}

func Start(threaded bool) {
	if threaded {
		var x C.int
		C.MPI_Init_thread(nil, nil, C.MPI_THREAD_MULTIPLE, &x)
		if x != C.MPI_THREAD_MULTIPLE {
			log.Fatalf("Requested threading support %d not available (%d).", C.MPI_THREAD_MULTIPLE, x)
		}
	} else {
		C.MPI_Init(nil, nil)
	}
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
	comm C.MPI_Comm
}

func WorldCommunicator() *Communicator {
	var o Communicator

	o.comm = C.MPI_COMM_WORLD
	return &o
}

func Size() int {
	var numProcesses C.int
	C.MPI_Comm_size(C.MPI_COMM_WORLD, &numProcesses)
	return int(numProcesses)
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

// ==================== Отправка и получение ======================

type Status struct {
	mpiStatus C.MPI_Status
}

func (s *Status) GetCount(t DataType) int {
	var n C.int
	C.MPI_Get_count(&(s.mpiStatus), dataTypes[t], &n)
	return int(n)
}

func (s *Status) GetTag() int {
	return int(s.mpiStatus.MPI_TAG)
}

func (s *Status) GetSource() int {
	return int(s.mpiStatus.MPI_SOURCE)
}

func (s *Status) GetError() int {
	return int(s.mpiStatus.MPI_ERROR)
}

func (o *Communicator) Iprobe(source int, tag int) (bool, *Status) {
	var s Status
	var b C.int

	C.MPI_Iprobe(C.int(source), C.int(tag), o.comm, &b, &(s.mpiStatus))
	return int(b) == 1, &s
}

func (o *Communicator) Probe(source int, tag int) *Status {
	var s Status
	C.MPI_Probe(C.int(source), C.int(tag), o.comm, &(s.mpiStatus))
	return &s
}

func (o *Communicator) SendBytes(vals []byte, toID int, tag int) {
	// log.Println("send:", "len=", vals, "\n", "toID=", toID, "tag=", tag)
	// if len(vals) > 1 && vals[len(vals) - 1] == 0 {
	// 	log.Panicln("hello!")
	// }

	var buf unsafe.Pointer
	if len(vals) == 0 {
		buf = unsafe.Pointer(nil)
	} else {
		buf = unsafe.Pointer(&vals[0])
	}
	C.MPI_Send(buf, C.int(len(vals)), dataTypes[Byte], C.int(toID), C.int(tag), o.comm)

}

var X_min = 0
var Len_min = 10000
var Len_max = 0

func (o *Communicator) RecvBytes(fromID int, tag int) ([]byte, Status) {
	l := o.Probe(fromID, tag).GetCount(Byte)
	_, st := o.Iprobe(fromID, tag)
	l_i := st.GetCount(Byte)
	if l != l_i {
		log.Println(l, l_i)
	}
	l_i = int(math.Max(float64(l), float64(l_i)))

	buf := make([]byte, l_i)
	cnt := 0
	status := o.RecvPreallocBytes(buf, len(buf), fromID, tag)
	for len(buf) > 0 && buf[len(buf)-1] == 0 {
		buf = buf[:len(buf)-1]
		cnt++
	}

	x := cnt - l - 1
	X_min = int(math.Min(float64(x), float64(X_min)))
	if len(buf) > 0 {
		Len_min = int(math.Min(float64(len(buf)), float64(Len_min)))
	}
	Len_max = int(math.Max(float64(len(buf)), float64(Len_max)))
	// log.Println(len(buf))
	// if (x == -4) {
	// 	log.Println(len(buf), buf, x)
	// }
	// log.Println("recv", "len=", buf, "\n", "fromID=", status.GetSource(), "tag=", tag)
	return buf, status
}

func (o *Communicator) RecvPreallocBytes(vals []byte, len int, fromID int, tag int) Status {
	buf := unsafe.Pointer(&vals[0])
	status := Status{}

	C.MPI_Recv(buf, C.int(len), dataTypes[Byte], C.int(fromID), C.int(tag), o.comm, &(status.mpiStatus))
	return status
}
