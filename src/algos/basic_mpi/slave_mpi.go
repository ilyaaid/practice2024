package basic_mpi

import (
	"CC/src/algos/algo_config"
	"CC/src/graph"
	"encoding/json"
	"fmt"
	"os"
	"path"

	mpi "github.com/sbromberger/gompi"
	"golang.org/x/sync/errgroup"
)

type Slave struct {
	// параметры для MPI
	rank       int
	comm       *mpi.Communicator
	slavesComm *mpi.Communicator

	// кофигурация для алгоритма
	conf *algo_config.AlgoConfig

	// параметры для части графа
	edges []graph.Edge

	changed  bool
	cc       map[graph.IndexType]graph.IndexType
	remotecc map[graph.IndexType]graph.IndexType

	ppNumber int
}

func (slave *Slave) IsOwnerOfVertex(v graph.IndexType) bool {
	return Vertex2Proc(slave.conf, v) == slave.rank
}

func (slave *Slave) Init() error {
	slave.changed = false
	slave.cc = make(map[graph.IndexType]graph.IndexType)
	slave.remotecc = make(map[graph.IndexType]graph.IndexType)
	return nil
}

// Получаем все необходимые ребра

func (slave *Slave) GetEdges() error {
	for {
		str, status := slave.comm.RecvString(MASTER_RANK, mpi.AnyTag)
		recvTag := status.GetTag()
		switch recvTag {
		case TAG_SEND_EDGE:
			edge, err := graph.StrToEdgeObj(str)
			if err != nil {
				return err
			}

			slave.edges = append(slave.edges, *edge)

			// добавление вершин в CC
			slave.cc[edge.V1] = edge.V1
			if slave.IsOwnerOfVertex(edge.V2) {
				slave.cc[edge.V2] = edge.V2
			} else {
				slave.remotecc[edge.V2] = edge.V2
			}
		case TAG_NEXT_PHASE:
			return nil
		default:
			return fmt.Errorf("slave.GetEdges: Wrong TAG %d", recvTag)
		}
	}
}

// Считаем количество получаемых сообщений на одном этапе

func (slave *Slave) countReceivingPPNumber() {
	countMp := make(map[graph.IndexType](map[graph.IndexType]bool))
	for _, edge := range slave.edges {
		if !slave.IsOwnerOfVertex(edge.V2) {
			_, isExist := countMp[edge.V1]
			if !isExist {
				countMp[edge.V1] = make(map[graph.IndexType]bool)
			}
			countMp[edge.V1][graph.IndexType(Vertex2Proc(slave.conf, edge.V2))] = true
		}
	}
	for _, mp := range countMp {
		slave.ppNumber += len(mp)
	}
}

// Далее все функции для алгоритма CC

func (slave *Slave) runHooking() {
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

func (slave *Slave) sendingPP() error {
	for remoteV, proposedParent := range slave.remotecc {
		remoteProc := Vertex2Proc(slave.conf, remoteV)
		slave.sendPP(remoteV, proposedParent, remoteProc)
	}
	return nil
}

func (slave *Slave) receivePP() (graph.IndexType, graph.IndexType, error) {
	mes, status := slave.slavesComm.RecvUint32s(mpi.AnySource, TAG_SEND_PP)
	if err := status.GetError(); err != 0 {
		return 0, 0, fmt.Errorf("slave.receivePP: MPI_ERROR=%d", err)
	}
	return graph.IndexType(mes[0]), graph.IndexType(mes[1]), nil
}

func (slave *Slave) receivingPP() error {
	for i := 0; i < slave.ppNumber; i++ {
		v, pp, err := slave.receivePP()
		if err != nil {
			return err
		}
		if slave.cc[v] > pp {
			slave.cc[v] = pp
			slave.changed = true
		}
	}
	return nil
}

func (slave *Slave) runPP() error {
	var g errgroup.Group
	g.SetLimit(2)
	g.Go(slave.receivingPP)
	g.Go(slave.sendingPP)
	
	if err := g.Wait(); err != nil {
		return err
	}

	if slave.changed {
		slave.comm.SendByte(byte(1), MASTER_RANK, TAG_IS_CHANGED)
	} else {
		slave.comm.SendByte(byte(0), MASTER_RANK, TAG_IS_CHANGED)
	}
	return nil
}

func (slave *Slave) CCSearch() error {
	for {
		slave.changed = false
		slave.runHooking()
		err := slave.runPP()
		if err != nil {
			return err
		}

		_, status := slave.comm.RecvString(MASTER_RANK, mpi.AnyTag)
		tag := status.GetTag()
		if tag == TAG_CONTINUE_CC {
			continue
		} else if tag == TAG_NEXT_PHASE { 
			break
		} else {
			return fmt.Errorf("slave.CCSearch: Wrong TAG %d from MASTER_RANK", tag)
		}
	}
	return nil
}

//=================== Получение результата

func (slave *Slave) addResult() error {
	//получаем путь к папке, общей для всех слейвов (для записи результата в нее)
	resDirPath, _ := slave.comm.RecvString(MASTER_RANK, TAG_SEND_RESULT_PATH)

	resFilePath := path.Join(resDirPath, fmt.Sprintf("slave_%d.txt", slave.rank))
	file, err := os.Create(resFilePath)
	if err != nil {
		return err
	}
	defer file.Close()

	bytes, err := json.Marshal(slave.cc)
	if err != nil {
		return err
	}
	file.WriteString(string(bytes))
	return nil
}

func (slave *Slave) sendResult() error {
	str, err := json.Marshal(slave.cc)
	if err != nil {
		return err
	}
	slave.comm.SendBytes(str, MASTER_RANK, TAG_SEND_RESULT)
	return nil
}
