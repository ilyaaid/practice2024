package basic_mpi

import (
	"CC/src/algos/algo_config"
	"CC/src/graph"
	"encoding/json"
	"log"

	mpi "github.com/sbromberger/gompi"
)

type Slave struct {
	// параметры для MPI
	rank int
	comm *mpi.Communicator
	slavesComm *mpi.Communicator

	// кофигурация для алгоритма
	conf *algo_config.AlgoConfig

	// параметры для части графа
	edges []graph.Edge

	changed bool
	cc map[graph.IndexType]graph.IndexType
	remotecc map[graph.IndexType]graph.IndexType

	ppNumber int
}

func (slave *Slave) IsOwnerOfVertex(v graph.IndexType) bool {
	return Vertex2Proc(slave.conf, v) == slave.rank
}

func (slave *Slave) Init() {
	slave.changed = false
	slave.cc = make(map[graph.IndexType]graph.IndexType)
	slave.remotecc = make(map[graph.IndexType]graph.IndexType)
}

// Получаем все необходимые ребра

func (slave *Slave) GetEdges() {
	for {
		str, status := slave.comm.RecvString(MASTER_RANK, mpi.AnyTag)
		recvTag := status.GetTag()
		switch recvTag {
		case TAG_SEND_EDGE:
			edge := graph.StrToEdgeObj(str)
			slave.edges = append(slave.edges, *edge)

			// добавление вершин в CC
			slave.cc[edge.V1] = edge.V1
			if slave.IsOwnerOfVertex(edge.V2) {
				slave.cc[edge.V2] = edge.V2
			} else {
				slave.remotecc[edge.V2] = edge.V2
			}
		case TAG_NEXT_PHASE:
			return
		default:
			log.Panicln("wrong tag in slave.getEdges")
		}
	}
}

// Считаем количество получаемых сообщений на одном этапе

func (slave *Slave) countReceivingPPNumber() {
	countMp := make(map[graph.IndexType](map[graph.IndexType]bool))
	cnt := 0
	for _, edge := range slave.edges {
		if !slave.IsOwnerOfVertex(edge.V2) {
			_, isExist := countMp[edge.V1]
			if !isExist {
				countMp[edge.V1] = make(map[graph.IndexType]bool)
			}
			countMp[edge.V1][graph.IndexType(Vertex2Proc(slave.conf, edge.V2))] = true
			cnt++
		}
	}
	// slave.ppNumber = cnt
	ppN := 0
	for _, mp := range countMp {
		ppN += len(mp)
	}
	slave.ppNumber = ppN
}


// Далее все функции для алгоритма CC

func (slave* Slave) runHooking() {
	for _, edge := range slave.edges {
		if slave.IsOwnerOfVertex(edge.V2) {
			if slave.cc[edge.V1] < slave.cc[edge.V2] {
				slave.cc[edge.V2] = slave.cc[edge.V1]
				slave.changed = true
			} else if slave.cc[edge.V1] > slave.cc[edge.V2] {
				slave.cc[edge.V1] = slave.cc[edge.V2]
				slave.changed = true
			}
		} else {
			if slave.cc[edge.V1] < slave.remotecc[edge.V2] {
				slave.remotecc[edge.V2] = slave.cc[edge.V1]
			} else if slave.cc[edge.V1] > slave.remotecc[edge.V2] {
				slave.cc[edge.V1] = slave.remotecc[edge.V2]
			}
		}
	}
}

func (slave *Slave) sendPP(remoteV graph.IndexType, proposedParent graph.IndexType, toProc int) {
	slave.slavesComm.SendUInt32s(
		[]uint32{uint32(remoteV), uint32(proposedParent)}, toProc, TAG_SEND_PP)
}

func (slave *Slave) sendingPP(exit chan bool) {
	for remoteV, proposedParent := range slave.remotecc {
		remoteProc := Vertex2Proc(slave.conf, remoteV)
		slave.sendPP(remoteV, proposedParent, remoteProc)
	}
	// log.Println("sending: exit")
	exit <- true
}

func (slave *Slave) receivePP() (graph.IndexType, graph.IndexType) {
	mes, _ := slave.slavesComm.RecvUint32s(mpi.AnySource, TAG_SEND_PP) 
	return graph.IndexType(mes[0]),graph.IndexType(mes[1])
}

func (slave *Slave) receivingPP(exit chan bool) {
	// log.Println("ppNumber:", slave.ppNumber)
	for i := 0; i < slave.ppNumber; i++ {
		v, pp := slave.receivePP()
		if slave.cc[v] > pp {
			slave.cc[v] = pp
			slave.changed = true
		}
	}
	// log.Println("receiving: exit")
	exit <- true
}

func (slave *Slave) runPP() {
	exit := make(chan bool)
	go slave.receivingPP(exit)
	go slave.sendingPP(exit)
	<-exit
	<-exit
	if (slave.changed) {
		slave.comm.SendByte(byte(1), MASTER_RANK, TAG_IS_CHANGED)
	} else {
		slave.comm.SendByte(byte(0), MASTER_RANK, TAG_IS_CHANGED)
	}
}

func (slave* Slave) CCSearch() {
	for {
		slave.changed = false
		slave.runHooking()
		slave.runPP()
		slave.slavesComm.Barrier()
		_, status := slave.comm.RecvString(MASTER_RANK, mpi.AnyTag)
		tag := status.GetTag()
		slave.slavesComm.Barrier()
		if tag == TAG_CONTINUE_CC {
			continue
		} else { // tag = TAG_NEXT_PHASE
			break
		}
	}
}

func (slave *Slave) sendResult() {
	str, err := json.Marshal(slave.cc)
	if err != nil {
		log.Panicln(err)
	}
	slave.comm.SendBytes(str, MASTER_RANK, TAG_SEND_RESULT)
}
