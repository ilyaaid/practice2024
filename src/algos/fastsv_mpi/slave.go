package fastsv_mpi

import (
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

	// объект самого алгоритма
	algo *Algo

	// параметры для части графа
	edges []graph.Edge

	fchanged bool
	ffchanged bool

	f     map[graph.IndexType]graph.IndexType
	ff    map[graph.IndexType]graph.IndexType
	fnext map[graph.IndexType]graph.IndexType
}

func (slave *Slave) IsOwnerOfVertex(v graph.IndexType) bool {
	return slave.algo.getSlave(v) == slave.rank
}

func (slave *Slave) Init() error {
	slave.fchanged = false
	slave.f = make(map[graph.IndexType]graph.IndexType)
	slave.fnext = make(map[graph.IndexType]graph.IndexType)
	slave.ff = make(map[graph.IndexType]graph.IndexType)
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
			slave.f[edge.V1] = edge.V1
			if slave.IsOwnerOfVertex(edge.V2) {
				slave.f[edge.V2] = edge.V2
			}
		case TAG_NEXT_PHASE:
			copyCC(slave.fnext, slave.f)
			copyCC(slave.ff, slave.f)
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

	iteration := 0

	slave.algo.logger.beginIterations()
	
	for !is_next_phase {
		slave.fchanged = false
		slave.ffchanged = false

		slave.algo.logger.beginStep(STEP_STOCH_H)
		err := step.run(STEP_STOCH_H)
		if err != nil {
			return err
		}
		slave.algo.logger.endStep()

		slave.slavesComm.Barrier()

		slave.algo.logger.beginStep(STEP_AGGR_H)
		err = step.run(STEP_AGGR_H)
		if err != nil {
			return err
		}
		slave.algo.logger.endStep()

		slave.slavesComm.Barrier()

		slave.algo.logger.beginStep(STEP_SHORTCUT_H)
		err = step.run(STEP_SHORTCUT_H)
		if err != nil {
			return err
		}
		slave.algo.logger.endStep()

		// если вариант запуска алгоритма с дискретным обновлением вектора
		if slave.algo.conf.Variant == VARIANT_DISC_F_PAR {
			copyCC(slave.f, slave.fnext)
		}

		// проверка на изменение родителей второго порядка
		changed := slave.fchanged
		if slave.algo.conf.Variant == VARIANT_CONT_S_PAR {
			changed = slave.fchanged && slave.ffchanged
		}

		if changed {
			mympi.SendTag(slave.comm, MASTER_RANK, TAG_IS_CHANGED)
		} else {
			mympi.SendTag(slave.comm, MASTER_RANK, TAG_IS_NOT_CHANGED)
		}

		status := mympi.RecvTag(slave.comm, MASTER_RANK, mympi.AnyTag)
		switch tag := status.GetTag(); tag {
		case TAG_CONTINUE_CC:
			slave.algo.logger.nextIteration()
			iteration++
			continue
		case TAG_NEXT_PHASE:
			slave.algo.logger.endIterations()
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

	bytes, err := json.Marshal(slave.f)
	if err != nil {
		return err
	}
	file.WriteString(string(bytes))
	return nil
}

func (slave *Slave) sendResult() error {
	str, err := json.Marshal(slave.f)
	if err != nil {
		return err
	}
	slave.comm.SendBytes(str, MASTER_RANK, TAG_SEND_RESULT)
	return nil
}
