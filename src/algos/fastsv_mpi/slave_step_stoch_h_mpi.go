package fastsv_mpi

import (
	"encoding/json"
	"log"
	"sync"

	mpi "github.com/sbromberger/gompi"
	"golang.org/x/sync/errgroup"
)

func (slave *Slave) receivingStochH(chainsCnt *int, mutex *sync.Mutex) error {
	for {
		log.Println("before recv")
		mesStr, status := slave.comm.RecvString(mpi.AnySource, mpi.AnyTag)
		tag := status.GetTag()
		if tag == TAG_END_STEP {
			log.Println("TAG_END_STEP")
			break
		}

		var mes MessageMPI
		err := json.Unmarshal([]byte(mesStr), &mes)
		if err != nil {
			return err
		}

		log.Println("recv mes=", mes, "from tag=", tag)

		switch tag {
		case TAG_STEP_0:
			mes.PPNonConst = slave.cc[mes.PPNonConst]
			slave.sendStochH(mes, TAG_STEP_1, chainsCnt, mutex)
		case TAG_STEP_1:
			mes.PPNonConst = slave.cc[mes.PPNonConst]
			slave.sendStochH(mes, TAG_STEP_2, chainsCnt, mutex)
		case TAG_STEP_2:
			if slave.ccnext[mes.V] > mes.PPNonConst {
				slave.ccnext[mes.V] = mes.PPNonConst
				slave.changed = true
			}
			slave.sendStochH(mes, TAG_STEP_3, chainsCnt, mutex)
		case TAG_STEP_3:
			mutex.Lock()
			*chainsCnt = *chainsCnt - 1
			if *chainsCnt == 0 {
				sendTag(slave.comm, MASTER_RANK, TAG_END_STEP)
				*chainsCnt = -1
			}
			mutex.Unlock()
		}
	}
	return nil
}

func (slave *Slave) sendStochH(mes MessageMPI, tag int, chainsCnt *int, mutex *sync.Mutex) error {
	log.Println("send mes=", mes, "by tag=", tag)
	switch tag {
	case TAG_STEP_0:
		if slave.IsOwnerOfVertex(mes.PPNonConst) {
			mes.PPNonConst = slave.cc[mes.PPNonConst]
			tag = TAG_STEP_1
		} else {
			toProc := Vertex2Proc(slave.conf, mes.PPNonConst)
			mesBytes, err := json.Marshal(mes)
			if err != nil {
				return err
			}
			slave.comm.SendString(string(mesBytes), toProc, tag)

			break
		}
		fallthrough
	case TAG_STEP_1:
		if slave.IsOwnerOfVertex(mes.PPNonConst) {
			mes.PPNonConst = slave.cc[mes.PPNonConst]
			tag = TAG_STEP_2
		} else {
			toProc := Vertex2Proc(slave.conf, mes.PPNonConst)
			mesBytes, err := json.Marshal(mes)
			if err != nil {
				return err
			}
			slave.comm.SendString(string(mesBytes), toProc, tag)
			break
		}
		fallthrough
	case TAG_STEP_2:
		if slave.IsOwnerOfVertex(mes.V) {
			if slave.ccnext[mes.V] > mes.PPNonConst {
				slave.ccnext[mes.V] = mes.PPNonConst
				slave.changed = true
			}
			tag = TAG_STEP_3
		} else {
			log.Println("send!!", mes)
			toProc := Vertex2Proc(slave.conf, mes.V)
			mesBytes, err := json.Marshal(mes)
			if err != nil {
				return err
			}
			slave.comm.SendString(string(mesBytes), toProc, tag)
			break
		}
		fallthrough
	case TAG_STEP_3:
		if slave.rank == mes.StartProc {
			mutex.Lock()
			*chainsCnt = *chainsCnt - 1
			if *chainsCnt == 0 {
				sendTag(slave.comm, MASTER_RANK, TAG_END_STEP)
				*chainsCnt = -1
			}
			mutex.Unlock()
		} else {
			mesBytes, err := json.Marshal(mes)
			if err != nil {
				return err
			}
			slave.comm.SendString(string(mesBytes), mes.StartProc, tag)
		}
	}
	return nil
}

func (slave *Slave) sendingStochH(chainsCnt *int, mutex *sync.Mutex) error {
	for _, edge := range slave.edges {
		mes := MessageMPI{V: slave.cc[edge.V1], PPNonConst: edge.V2, StartProc: slave.rank}
		err := slave.sendStochH(mes, TAG_STEP_0, chainsCnt, mutex)
		if err != nil {
			return err
		}
	}
	return nil
}

func (slave *Slave) runStochH() error {
	var g errgroup.Group
	g.SetLimit(2)
	chainsCnt := len(slave.edges)

	if chainsCnt == 0 {
		sendTag(slave.comm, MASTER_RANK, TAG_END_STEP)
		_ = recvTag(slave.comm, MASTER_RANK, TAG_END_STEP)
	} else {
		mutex := &sync.Mutex{}
		slave.sendingStochH(&chainsCnt, mutex)
		slave.receivingStochH(&chainsCnt, mutex)

		// g.Go(func() error {
		// 	return slave.sendingStochH(&chainsCnt, mutex)
		// })
		// g.Go(func() error {
		// 	return slave.receivingStochH(&chainsCnt, mutex)
		// })
	}

	if err := g.Wait(); err != nil {
		return err
	}
	return nil
}
