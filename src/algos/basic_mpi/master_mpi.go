package basic_mpi

import (
	"CC/src/algos/algo_config"
	"CC/src/graph"
	"encoding/json"
	"fmt"
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
	master.BcastSend(" ", TAG_NEXT_PHASE)
	return nil
}

func (master *Master) BcastSend(message string, tag int) {
	for i := 0; i < master.conf.ProcNum; i++ {
		master.comm.SendString(message, i, tag)
	}
}

// Управление алгоритмом CC
func (master *Master) manageCCSearch() error {
	changed := true
	for changed {
		changed = false
		for i := 0; i < master.conf.ProcNum; i++ {
			mes, status := master.comm.RecvByte(i, TAG_IS_CHANGED)
			if err := status.GetError(); err != 0 {
				return fmt.Errorf("master.manageCCSearch: MPI_ERROR %d", err)
			}

			is_changed := mes != 0
			if is_changed {
				changed = is_changed
			}
		}
		if changed {
			master.BcastSend(" ", TAG_CONTINUE_CC)
		}
	}
	master.BcastSend(" ", TAG_NEXT_PHASE)
	return nil
}

// Результат

func (master *Master) getResult() error {
	for i := 0; i < master.conf.ProcNum; i++ {
		mes, status := master.comm.RecvBytes(i, TAG_SEND_RESULT)
		if err := status.GetError(); err != 0 {
			return fmt.Errorf("master.getResult: MPI_ERROR %d", err)
		}

		var cc map[graph.IndexType]graph.IndexType
		err := json.Unmarshal(mes, &cc)
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
