package fastsv_mpi

import (
	"CC/algos/algo_config"
	"CC/graph"
	"encoding/json"
	"fmt"
	// "log"
	"os"
	"path"

	mpi "github.com/sbromberger/gompi"
)

var MASTER_RANK int = 0

type Master struct {
	// параметры для MPI
	comm *mpi.Communicator

	// кофигурация для алгоритма
	conf *algo_config.AlgoConfig

	// параметры для графа
	g *graph.Graph
}

func (master *Master) Init() error {
	var err error
	master.g, err = master.conf.GrIO.Read()
	if err != nil {
		return err
	}
	return nil
}

// распределение ребер по процессам
func (master *Master) SendAllEdges() error {
	for _, edge := range master.g.Edges {
		proc1, proc2 := Vertex2Proc(master.conf, edge.V1), Vertex2Proc(master.conf, edge.V2)
		edgeStr, err := edge.ToStr()
		if err != nil {
			return err
		}

		master.comm.SendString(edgeStr, proc1, TAG_SEND_EDGE)
		if proc1 != proc2 {
			edge.V1, edge.V2 = edge.V2, edge.V1
			edgeStr, err := edge.ToStr()
			if err != nil {
				return err
			}
			master.comm.SendString(edgeStr, proc2, TAG_SEND_EDGE)
		}
	}
	master.BcastSendTag(TAG_NEXT_PHASE)
	return nil
}

func (master *Master) BcastSend(message string, tag int) {
	for i := 0; i < master.conf.ProcNum; i++ {
		master.comm.SendString(message, i, tag)
	}
}

func (master *Master) BcastSendTag(tag int) {
	master.BcastSend(" ", tag)
}

// Управление алгоритмом CC
func (master *Master) manageCCSearch() error {
	changed := true

	for changed {
		for step := 0; step < 1; step++ {
			end_step_cnt := 0
			for i := 0; i < master.conf.ProcNum; i++ {
				// log.Println("wait")
				_ = recvTag(master.comm, mpi.AnySource, TAG_END_STEP)
				// log.Println("recv from", status.GetSource())
				end_step_cnt++
			}
			if end_step_cnt == master.conf.ProcNum {
				master.BcastSendTag(TAG_END_STEP)
			}
		}

		changed = false
		for i := 0; i < master.conf.ProcNum; i++ {
			status := recvTag(master.comm, i, mpi.AnyTag)
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
			master.BcastSendTag(TAG_CONTINUE_CC)
		}
	}
	master.BcastSendTag(TAG_NEXT_PHASE)
	return nil
}

// Результат

func (master *Master) getResult() error {
	for i := 0; i < master.conf.ProcNum; i++ {
		mes, status := master.comm.RecvString(i, TAG_SEND_RESULT)
		if err := status.GetError(); err != 0 {
			return fmt.Errorf("master.getResult: MPI_ERROR %d", err)
		}

		var cc map[graph.IndexType]graph.IndexType
		err := json.Unmarshal([]byte(mes), &cc)
		if err != nil {
			return err
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
	master.BcastSend(resDirPath, TAG_SEND_RESULT_PATH)
	return nil
}
