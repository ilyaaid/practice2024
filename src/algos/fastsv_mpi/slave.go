package fastsv_mpi

import (
	"CC/algos/algo_config"
	"CC/graph"
	"CC/mympi"
	"encoding/json"
	"fmt"
	"os"
	"path"
)

type Slave struct {
	// параметры для MPI
	rank       int
	comm       *mympi.Communicator
	slavesComm *mympi.Communicator

	// кофигурация для алгоритма
	conf *algo_config.AlgoConfig

	// параметры для части графа
	edges []graph.Edge

	changed bool
	cc      map[graph.IndexType]graph.IndexType
	ccnext  map[graph.IndexType]graph.IndexType
}

func (slave *Slave) IsOwnerOfVertex(v graph.IndexType) bool {
	return Vertex2Proc(slave.conf, v) == slave.rank
}

func (slave *Slave) Init() error {
	slave.changed = false
	slave.cc = make(map[graph.IndexType]graph.IndexType)
	slave.ccnext = make(map[graph.IndexType]graph.IndexType)
	return nil
}

// Получаем все необходимые ребра

func (slave *Slave) GetEdges() error {
	for {
		edgeBytes, status := slave.comm.RecvBytes(MASTER_RANK, mympi.AnyTag)
		recvTag := status.GetTag()
		switch recvTag {
		case TAG_SEND_EDGE:
			edge, err := graph.StrToEdgeObj(edgeBytes)
			if err != nil {
				return err
			}

			slave.edges = append(slave.edges, *edge)

			// добавление вершин в CC
			slave.cc[edge.V1] = edge.V1
			slave.ccnext[edge.V1] = edge.V1
			if slave.IsOwnerOfVertex(edge.V2) {
				slave.cc[edge.V2] = edge.V2
				slave.ccnext[edge.V2] = edge.V2
			}
		case TAG_NEXT_PHASE:
			return nil
		default:
			return fmt.Errorf("slave.GetEdges: Wrong TAG %d", recvTag)
		}
	}
}

// Далее все функции для алгоритма CC
func (slave *Slave) runSteps() error {
	is_next_phase := false
	step := Step{slave: slave}
	step.Init()

	for !is_next_phase {
		slave.changed = false

		err := step.run(STEP_STOCH_H)
		if err != nil {
			return err
		}

		slave.slavesComm.Barrier()

		err = step.run(STEP_AGGR_H)
		if err != nil {
			return err
		}

		slave.slavesComm.Barrier()

		err = step.run(STEP_SHORTCUT_H)
		if err != nil {
			return err
		}

		copyCC(slave.cc, slave.ccnext)

		if slave.changed {
			mympi.SendTag(slave.comm, MASTER_RANK, TAG_IS_CHANGED)
		} else {
			mympi.SendTag(slave.comm, MASTER_RANK, TAG_IS_NOT_CHANGED)
		}

		status := mympi.RecvTag(slave.comm, MASTER_RANK, mympi.AnyTag)
		switch tag := status.GetTag(); tag {
		case TAG_CONTINUE_CC:
			continue
		case TAG_NEXT_PHASE:
			is_next_phase = true
		default:
			return fmt.Errorf("runSteps: wrong TAG %d", tag)
		}
	}
	return nil
}

//=================== Получение результата

func (slave *Slave) addResult() error {
	//получаем путь к папке, общей для всех слейвов (для записи результата в нее)
	resDirPath, _ := slave.comm.RecvBytes(MASTER_RANK, TAG_SEND_RESULT_PATH)

	resFilePath := path.Join(string(resDirPath), fmt.Sprintf("slave_%d.txt", slave.rank))
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
