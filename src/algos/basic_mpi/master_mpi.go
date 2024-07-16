package basic_mpi

import (
	"CC/src/algos/algo_config"
	"CC/src/graph"
	"encoding/json"
	"log"

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

func (master *Master) Init() {
	master.g = graph.ReadFromFile(master.conf.GraphFilename)
}

// распределение ребер по процессам
func (master *Master) SendAllEdges() {
	for _, edge := range master.g.Edges {
		proc1, proc2 := Vertex2Proc(master.conf, edge.V1), Vertex2Proc(master.conf, edge.V2)
		master.comm.SendString(edge.ToStr(), proc1, TAG_SEND_EDGE)
		if proc1 != proc2 {
			edge.V1, edge.V2 = edge.V2, edge.V1
			master.comm.SendString(edge.ToStr(), proc2, TAG_SEND_EDGE)
		}
	}
	master.BcastSend(" ", TAG_NEXT_PHASE)
}


func (master *Master) BcastSend(message string, tag int) {
	for i := 0; i < master.conf.ProcNum; i++ {
		master.comm.SendString(message, i, tag)
	}
}

func (master *Master) manageCCSearch() {
	changed := true
	for changed {
		changed = false
		for i := 0; i < master.conf.ProcNum; i++ {
			mes, _ := master.comm.RecvByte(i, TAG_IS_CHANGED)
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
}


func (master *Master) getResult() {
	for i := 0; i < master.conf.ProcNum; i++ {
		mes, _ := master.comm.RecvBytes(i, TAG_SEND_RESULT)
		var cc map[graph.IndexType]graph.IndexType
		err := json.Unmarshal(mes, &cc)
		if err != nil {
			log.Panicln(err)
		}
		log.Println(cc)
	}
}
