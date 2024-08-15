package basic_mpi

import (
	"CC/graph"
	"CC/mympi"
	"encoding/json"
	"fmt"
	"sync"

	"golang.org/x/sync/errgroup"
)

type Slave struct {
	// параметры для MPI
	rank       int
	comm       *mympi.Communicator
	slavesComm *mympi.Communicator

	algo *Algo

	// параметры для части графа
	edges []graph.Edge

	changed  bool
	f       map[graph.IndexType]graph.IndexType
	remotef map[graph.IndexType]graph.IndexType

	// fnext map[graph.IndexType]graph.IndexType

	ppNumber int
}

func (slave *Slave) IsOwnerOfVertex(v graph.IndexType) bool {
	return slave.algo.getSlave(v) == slave.rank
}

func (slave *Slave) Init() error {
	slave.changed = false
	slave.f = make(map[graph.IndexType]graph.IndexType)
	slave.remotef = make(map[graph.IndexType]graph.IndexType)
	return nil
}

// Получаем все необходимые ребра

func (slave *Slave) GetEdges() error {
	for {
		str, status := slave.comm.RecvBytes(MASTER_RANK, mympi.AnyTag)
		if tag := status.GetTag(); tag == TAG_SEND_EDGE {
			edge, err := graph.StrToEdgeObj(str)
			if err != nil {
				return err
			}

			slave.edges = append(slave.edges, *edge)
			// добавление вершин в CC
			slave.f[edge.V1] = edge.V1
			if slave.IsOwnerOfVertex(edge.V2) {
				slave.f[edge.V2] = edge.V2
			} else {
				slave.remotef[edge.V2] = edge.V2
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
			countMp[edge.V1][graph.IndexType(slave.algo.getSlave(edge.V2))] = true
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
			if slave.f[edge.V1] < slave.f[edge.V2] {
				slave.f[edge.V2] = slave.f[edge.V1]
				slave.changed = true
			} else if slave.f[edge.V1] > slave.f[edge.V2] {
				slave.f[edge.V1] = slave.f[edge.V2]
				slave.changed = true
			}
		} else {
			if slave.f[edge.V1] < slave.remotef[edge.V2] {
				slave.remotef[edge.V2] = slave.f[edge.V1]
			} else if slave.f[edge.V1] > slave.remotef[edge.V2] {
				slave.f[edge.V1] = slave.remotef[edge.V2]
			}
		}
	}
}

type PPnode struct {
	RemoteV graph.IndexType
	PP      graph.IndexType
}

func (slave *Slave) sendPP(mutex *sync.Mutex, ppnode *PPnode, toID int) error {
	ppnodeBytes, err := json.Marshal(ppnode)
	if err != nil {
		return err
	}
	mutex.Lock()
	slave.comm.SendBytes(ppnodeBytes, toID, TAG_SEND_PP)
	mutex.Unlock()

	slave.algo.logger.sendTag()

	return nil
}

func (slave *Slave) sendingPP(mutex *sync.Mutex) error {
	for remoteV, proposedParent := range slave.remotef {
		remoteProc := slave.algo.getSlave(remoteV)
		slave.sendPP(mutex, &PPnode{remoteV, proposedParent}, remoteProc)
	}
	return nil
}

func (slave *Slave) receivePP(mutex *sync.Mutex) (*PPnode, error) {
	var mes []byte
	for {
		mutex.Lock()
		is_exist, _ := slave.comm.Iprobe(mympi.AnySource, TAG_SEND_PP)
		if is_exist {
			mes, _ = slave.comm.RecvBytes(mympi.AnySource, TAG_SEND_PP)
			mutex.Unlock()
			break
		}
		mutex.Unlock()
	}

	slave.algo.logger.recvTag()

	var ppnode PPnode
	err := json.Unmarshal(mes, &ppnode)
	if err != nil {
		return nil, err
	}
	return &ppnode, nil
}

func (slave *Slave) receivingPP(mutex *sync.Mutex) error {
	for i := 0; i < slave.ppNumber; i++ {
		ppnode, err := slave.receivePP(mutex)
		if err != nil {
			return err
		}
		if slave.f[ppnode.RemoteV] > ppnode.PP {
			slave.f[ppnode.RemoteV] = ppnode.PP
			slave.changed = true
		}
	}
	return nil
}

func (slave *Slave) runPP() error {
	var g errgroup.Group
	g.SetLimit(2)

	mutex := &sync.Mutex{}

	g.Go(func() error {
		return slave.sendingPP(mutex)
	})
	g.Go(func() error {
		return slave.receivingPP(mutex)
	})

	// slave.sendingPP(mutex)
	// slave.receivePP(mutex)

	if err := g.Wait(); err != nil {
		return err
	}
	return nil
}

func (slave *Slave) CCSearch() error {
	slave.algo.logger.beginIterations()
	for {
		slave.algo.logger.beginIteration()

		slave.changed = false
		slave.runHooking()
		err := slave.runPP()
		if err != nil {
			return err
		}

		slave.algo.logger.endIteration()

		if slave.changed {
			mympi.SendTag(slave.comm, MASTER_RANK, TAG_IS_CHANGED)
		} else {
			mympi.SendTag(slave.comm, MASTER_RANK, TAG_IS_NOT_CHANGED)
		}

		status := mympi.RecvTag(slave.comm, MASTER_RANK, mympi.AnyTag)
		if tag := status.GetTag(); tag == TAG_CONTINUE_CC {
			continue
		} else if tag == TAG_NEXT_PHASE {
			break
		} else {
			return fmt.Errorf("wrong tag=%d from MASTER_RANK", tag)
		}
	}

	slave.algo.logger.endIterations()
	return nil
}

//=================== Получение результата

func (slave *Slave) sendResult() error {
	str, err := json.Marshal(slave.f)
	if err != nil {
		return err
	}
	slave.comm.SendBytes(str, MASTER_RANK, TAG_SEND_RESULT)
	return nil
}

// func (slave *Slave) addResult() error {
// 	//получаем путь к папке, общей для всех слейвов (для записи результата в нее)
// 	resDirPath, _ := slave.comm.RecvString(MASTER_RANK, TAG_SEND_RESULT_PATH)

// 	resFilePath := path.Join(resDirPath, fmt.Sprintf("slave_%d.txt", slave.rank))
// 	file, err := os.Create(resFilePath)
// 	if err != nil {
// 		return err
// 	}
// 	defer file.Close()

// 	bytes, err := json.Marshal(slave.cc)
// 	if err != nil {
// 		return err
// 	}
// 	file.WriteString(string(bytes))
// 	return nil
// }
