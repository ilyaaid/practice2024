package fastsv_mpi

import (
	"CC/graph"
	"CC/mympi"
	"encoding/json"
	"fmt"
	"os"
	"path"
)

var MASTER_RANK int = 0

type Master struct {
	// параметры для MPI
	comm *mympi.Communicator

	algo *Algo

	// параметры для графа
	g *graph.Graph
}

func (master *Master) Init() error {
	var err error
	master.g, err = master.algo.conf.GrIO.Read()
	if err != nil {
		return err
	}
	return nil
}

// распределение ребер по процессам
func (master *Master) SendAllEdges() error {
	for _, edge := range master.g.Edges {
		proc1, proc2 := master.algo.getSlave(edge.V1), master.algo.getSlave(edge.V2)
		edgeBytes, err := edge.ToBytes()
		if err != nil {
			return err
		}

		master.comm.SendBytes(edgeBytes, proc1, TAG_SEND_EDGE)
		if proc1 != proc2 {
			edge.V1, edge.V2 = edge.V2, edge.V1
			edgeBytes, err := edge.ToBytes()
			if err != nil {
				return err
			}
			master.comm.SendBytes(edgeBytes, proc2, TAG_SEND_EDGE)
		}
	}
	mympi.BcastSendTag(MASTER_RANK, TAG_NEXT_PHASE)
	return nil
}

// Управление алгоритмом CC
func (master *Master) manageCCSearch() error {
	changed := true

	for changed {
		for step := 0; step < 3; step++ {
			end_step_cnt := 0
			for i := 0; i < master.algo.conf.ProcNum; i++ {
				_ = mympi.RecvTag(master.comm, mympi.AnySource, TAG_END_STEP)
				end_step_cnt++
			}
			if end_step_cnt == master.algo.conf.ProcNum {
				mympi.BcastSendTag(MASTER_RANK, TAG_END_STEP)
			}
		}

		changed = false
		for i := 0; i < master.algo.conf.ProcNum; i++ {
			status := mympi.RecvTag(master.comm, mympi.AnySource, mympi.AnyTag)
			switch tag := status.GetTag(); tag {
			case TAG_IS_CHANGED:
				changed = true
			case TAG_IS_NOT_CHANGED:
				continue
			default:
				return fmt.Errorf("master.manageCCSearch(changed): Wrong TAG=%d", tag)
			}
		}
		if changed {
			mympi.BcastSendTag(MASTER_RANK, TAG_CONTINUE_CC)
		}
		master.algo.logger.nextIteration()
	}
	mympi.BcastSendTag(MASTER_RANK, TAG_NEXT_PHASE)
	return nil
}

// Результат

func (master *Master) getResult() error {
	
	for i := 0; i < master.algo.conf.ProcNum; i++ {
		mes, _ := master.comm.RecvBytes(i, TAG_SEND_RESULT)

		var cc map[graph.IndexType]graph.IndexType
		err := json.Unmarshal(mes, &cc)
		if err != nil {
			return err
		}

		for v, vcc := range cc {
			master.g.CC[v] = vcc
		}
	}
	
	return nil
}

func (master *Master) prepResult() error {
	//TODO доделать генерацию правильного нового пути для результата
	resDirPath := path.Join(".", "result")

	err := os.Mkdir(resDirPath, 0755)
	if err != nil {
		return fmt.Errorf("master.prepResult: folder creation error")
	}
	mympi.BcastSendBytes([]byte(resDirPath), MASTER_RANK, TAG_SEND_RESULT_PATH)
	return nil
}
