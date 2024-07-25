package basic_mpi

import (
	"CC/algos/algo_config"
	"CC/algos/algo_types"
	"CC/graph"
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

func (master *Master) BcastSend(mes []byte, tag int) {
	for i := 0; i < master.conf.ProcNum; i++ {
		sendMes(master.comm, mes, i, tag)
	}
}

func (master *Master) BcastSendTag(tag int) {
	for i := 0; i < master.conf.ProcNum; i++ {
		sendTag(master.comm, i, tag)
	}
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
		edgeBytes, err := edge.ToBytes()
		if err != nil {
			return err
		}

		sendMes(master.comm, edgeBytes, proc1, TAG_SEND_EDGE)
		if proc1 != proc2 {
			edge.V1, edge.V2 = edge.V2, edge.V1
			edgeBytes, err := edge.ToBytes()
			if err != nil {
				return err
			}
			sendMes(master.comm, edgeBytes, proc2, TAG_SEND_EDGE)
		}
	}
	master.BcastSendTag(TAG_NEXT_PHASE)
	return nil
}

// Управление алгоритмом CC
func (master *Master) manageCCSearch() error {
	changed := true
	for changed {
		changed = false
		for i := 0; i < master.conf.ProcNum; i++ {
			status := recvTag(master.comm, i, mpi.AnyTag)
			if tag := status.GetTag(); 
			tag == TAG_IS_CHANGED  {
				changed = true
			} else if tag == TAG_IS_NOT_CHANGED {
				continue
			} else {
				return fmt.Errorf("wrong TAG =%d", tag)
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

func CreateDir(path string) error {
	_, err := os.Stat(path)
	if os.IsNotExist(err) {
		err := os.Mkdir(path, 0755)
		if err != nil {
			return fmt.Errorf("folder %s creation error", path)
		}
	}
	return nil
}

func (master *Master) prepResult() error {
	//TODO доделать генерацию правильного нового пути для результата
	resDirPath := path.Join(".", "result")
	err := CreateDir(resDirPath)
	if err != nil {
		return err
	}
	resDirAlgoPath := path.Join(resDirPath, algo_types.ALGO_basic_mpi)
	err = CreateDir(resDirAlgoPath)
	if err != nil {
		return err
	}

	master.BcastSend([]byte(resDirAlgoPath), TAG_SEND_RESULT_PATH)
	return nil
}
