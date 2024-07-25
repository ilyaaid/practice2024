package basic_mpi

import (
	"CC/algos/algo_config"
	"CC/graph"
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
		str, status := recvMes(slave.comm, MASTER_RANK, mpi.AnyTag)
		if tag := status.GetTag();
		tag == TAG_SEND_EDGE {
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
		} else if tag == TAG_NEXT_PHASE {
			break
		} else {
			return fmt.Errorf("master.GetEdges: wrong TAG=%v", tag)
		}
	}
	return nil
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

type PPnode struct {
	RemoteV graph.IndexType
	PP      graph.IndexType
}

func (slave *Slave) sendPP(ppnode *PPnode, toID int) error {
	ppnodeBytes, err := json.Marshal(ppnode)
	if err != nil {
		return err
	}
	sendMes(slave.comm, ppnodeBytes, toID, TAG_SEND_PP)
	// log.Println("send:", ppnodeBytes, toID)
	return nil
}

func (slave *Slave) sendingPP() error {
	for remoteV, proposedParent := range slave.remotecc {
		remoteProc := Vertex2Proc(slave.conf, remoteV)
		slave.sendPP(&PPnode{remoteV, proposedParent}, remoteProc)
	}
	return nil
}

func (slave *Slave) receivePP() (*PPnode, error) {
	mes, _ := recvMes(slave.comm, mpi.AnySource, TAG_SEND_PP)
	// log.Println("recv:", mes, status.GetSource(), status.GetTag())

	var ppnode PPnode
	err := json.Unmarshal(mes, &ppnode)
	if err != nil {
		return nil, err
	}
	return &ppnode, nil
}

func (slave *Slave) receivingPP() error {
	for i := 0; i < slave.ppNumber; i++ {
		ppnode, err := slave.receivePP()
		if err != nil {
			return err
		}
		if slave.cc[ppnode.RemoteV] > ppnode.PP {
			slave.cc[ppnode.RemoteV] = ppnode.PP
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

		if slave.changed {
			sendTag(slave.comm, MASTER_RANK, TAG_IS_CHANGED)
		} else {
			sendTag(slave.comm, MASTER_RANK, TAG_IS_NOT_CHANGED)
		}

		status := recvTag(slave.comm, MASTER_RANK, mpi.AnyTag)
		if tag := status.GetTag(); 
		tag == TAG_CONTINUE_CC {
			continue
		} else if tag == TAG_NEXT_PHASE {
			break
		} else {
			return fmt.Errorf("wrong tag=%d from MASTER_RANK", tag)
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
