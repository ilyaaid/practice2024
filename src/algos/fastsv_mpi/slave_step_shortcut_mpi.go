package fastsv_mpi

import (
	"encoding/json"
	"sync"

	mpi "github.com/sbromberger/gompi"
	"golang.org/x/sync/errgroup"
)

func (slave *Slave) receivingShortcutH(chainsCnt *int, mutex *sync.Mutex) error {
	for {
		mesStr, status := slave.comm.RecvString(mpi.AnySource, mpi.AnyTag)
		tag := status.GetTag()
		if tag == TAG_END_STEP {
			break
		}

		var mes MessageMPI
		err := json.Unmarshal([]byte(mesStr), &mes)
		if err != nil {
			return err
		}

		switch tag {
		case TAG_STEP_0:
			mes.PPNonConst = slave.cc[mes.PPNonConst]
			slave.sendShortcutH(&mes, TAG_STEP_1, chainsCnt, mutex)
		case TAG_STEP_1:
			if slave.ccnext[mes.V] > mes.PPNonConst {
				slave.ccnext[mes.V] = mes.PPNonConst
				slave.changed = true
			}
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

func (slave *Slave) sendShortcutH(mes *MessageMPI, tag int, chainsCnt *int, mutex *sync.Mutex) error {
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
		startProc := Vertex2Proc(slave.conf, mes.V)
		if slave.rank == startProc {
			if slave.ccnext[mes.V] > mes.PPNonConst {
				slave.ccnext[mes.V] = mes.PPNonConst
				slave.changed = true
			}
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
			slave.comm.SendString(string(mesBytes), startProc, tag)
		}
	}
	return nil
}

func (slave *Slave) sendingShortcutH(chainsCnt *int, mutex *sync.Mutex) error {
	for _, edge := range slave.edges {
		mes := MessageMPI{V: edge.V1, PPNonConst: slave.cc[edge.V1]}
		err := slave.sendShortcutH(&mes, TAG_STEP_0, chainsCnt, mutex)
		if err != nil {
			return err
		}
	}
	return nil
}

func (slave *Slave) runShortcutH() error {
	var g errgroup.Group
	g.SetLimit(2)
	chainsCnt := len(slave.edges)
	mutex := &sync.Mutex{}
	g.Go(func() error { return slave.receivingShortcutH(&chainsCnt, mutex) })
	g.Go(func() error { return slave.sendingShortcutH(&chainsCnt, mutex) })

	if err := g.Wait(); err != nil {
		return err
	}
	return nil
}
