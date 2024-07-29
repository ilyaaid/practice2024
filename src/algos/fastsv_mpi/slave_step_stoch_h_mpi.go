package fastsv_mpi

import (
	"CC/mympi"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"sync"

	"golang.org/x/sync/errgroup"
)

func (slave *Slave) reduceChains(mutex *sync.Mutex) {
	mutex.Lock()
	slave.chainsCnt = slave.chainsCnt - 1
	if slave.chainsCnt == 0 {
		mympi.SendTag(slave.comm, MASTER_RANK, TAG_END_STEP)
	}
	mutex.Unlock()
}

func (slave *Slave) receivingStochH(ctx *context.Context, mutex *sync.Mutex) error {
	for {
		var mes []byte
		var status mympi.Status
		
		is_exist := false
		for {
			mutex.Lock()
			is_exist, _ = slave.comm.Iprobe(mympi.AnySource, mympi.AnyTag)
			if is_exist {
				mes, status = slave.comm.RecvBytes(mympi.AnySource, mympi.AnyTag)
				mutex.Unlock()
				break
			}
			mutex.Unlock()
		}

		if err := status.GetError(); err != 0 {
			log.Println("recv status error:", err)
			return fmt.Errorf("recv status error:%v", err)
		}

		tag := status.GetTag()
		if tag == TAG_END_STEP {
			break
		}

		var mesObj MessageMPI

		err := json.Unmarshal(mes, &mesObj)
		if err != nil {
			return fmt.Errorf("receivingStochH:%v", err)
		}

		switch tag {
		case TAG_STEP_0:
			mutex.Lock()
			mesObj.PPNonConst = slave.cc[mesObj.PPNonConst]
			mutex.Unlock()
			err := slave.sendStochH(mesObj, TAG_STEP_1, mutex)
			if err != nil {
				return fmt.Errorf("send TAG_STEP_1: %v", err)
			}
		case TAG_STEP_1:
			mutex.Lock()
			mesObj.PPNonConst = slave.cc[mesObj.PPNonConst]
			mutex.Unlock()
			err := slave.sendStochH(mesObj, TAG_STEP_2, mutex)
			if err != nil {
				return fmt.Errorf("send TAG_STEP_2: %v", err)
			}
		case TAG_STEP_2:
			mutex.Lock()
			if slave.ccnext[mesObj.V] > mesObj.PPNonConst {
				slave.ccnext[mesObj.V] = mesObj.PPNonConst
				slave.changed = true
			}
			mutex.Unlock()
			err := slave.sendStochH(mesObj, TAG_STEP_3, mutex)
			if err != nil {
				return fmt.Errorf("send TAG_STEP_3: %v", err)
			}
		case TAG_STEP_3:
			if mesObj.StartProc != slave.rank {
				log.Panicln("!!!!!!!!!!!!!!!", mesObj, slave.rank)
			}
			slave.reduceChains(mutex)
		default:
			return fmt.Errorf("wrong TAG in receivingStochH: %d", tag)
		}
	}
	return nil
}

func (slave *Slave) sendStochH(mes MessageMPI, tag int, mutex *sync.Mutex) error {
	switch tag {
	case TAG_STEP_0:
		toProc := Vertex2Proc(slave.conf, mes.PPNonConst)

		if toProc == slave.rank {
			mutex.Lock()
			mes.PPNonConst = slave.cc[mes.PPNonConst]
			mutex.Unlock()
			tag = TAG_STEP_1
		} else {
			mesBytes, err := json.Marshal(mes)
			if err != nil {
				return err
			}
			mutex.Lock()
			slave.comm.SendBytes(mesBytes, toProc, tag)
			mutex.Unlock()

			return nil
		}
		fallthrough
	case TAG_STEP_1:
		toProc := Vertex2Proc(slave.conf, mes.PPNonConst)
		if slave.rank == toProc {
			mutex.Lock()
			mes.PPNonConst = slave.cc[mes.PPNonConst]
			mutex.Unlock()
			tag = TAG_STEP_2
		} else {
			mesBytes, err := json.Marshal(mes)
			if err != nil {
				return err
			}
			mutex.Lock()
			slave.comm.SendBytes(mesBytes, toProc, tag)
			mutex.Unlock()

			return nil
		}
		fallthrough
	case TAG_STEP_2:
		toProc := Vertex2Proc(slave.conf, mes.V)
		if toProc == slave.rank {
			mutex.Lock()
			if slave.ccnext[mes.V] > mes.PPNonConst {
				slave.ccnext[mes.V] = mes.PPNonConst
				slave.changed = true
			}
			mutex.Unlock()
			tag = TAG_STEP_3
		} else {
			mesBytes, err := json.Marshal(mes)
			if err != nil {
				return err
			}
			mutex.Lock()
			slave.comm.SendBytes(mesBytes, toProc, tag)
			mutex.Unlock()

			return nil
		}
		fallthrough
	case TAG_STEP_3:
		if slave.rank == mes.StartProc {
			slave.reduceChains(mutex)
		} else {
			mesBytes, err := json.Marshal(mes)
			if err != nil {
				return err
			}
			mutex.Lock()
			slave.comm.SendBytes(mesBytes, mes.StartProc, tag)
			mutex.Unlock()

			return nil
		}
	}
	return nil
}

func (slave *Slave) sendingStochH(ctx *context.Context, mutex *sync.Mutex) error {
	for _, edge := range slave.edges {
		mutex.Lock()
		mes := MessageMPI{V: slave.cc[edge.V1], PPNonConst: edge.V2, StartProc: slave.rank}
		mutex.Unlock()
		err := slave.sendStochH(mes, TAG_STEP_0, mutex)
		if err != nil {
			return fmt.Errorf("sendingStoch: %v", err)
		}

	}
	return nil
}

func (slave *Slave) runStochH() error {
	g, ctx := errgroup.WithContext(context.Background())
	g.SetLimit(2)

	slave.chainsCnt = len(slave.edges)
	if slave.chainsCnt == 0 {
		mympi.SendTag(slave.comm, MASTER_RANK, TAG_END_STEP)
		_ = mympi.RecvTag(slave.comm, MASTER_RANK, TAG_END_STEP)
	} else {
		mutex := &sync.Mutex{}
		g.Go(func() error {
			return slave.sendingStochH(&ctx, mutex)
		})
		g.Go(func() error {
			return slave.receivingStochH(&ctx, mutex)
		})

		// slave.sendingStochH(&ctx, mutex)
		// slave.receivingStochH(&ctx, mutex)
	}

	if err := g.Wait(); err != nil {
		log.Panicln("!!!!!!wait:", err)
		return err
	}
	return nil
}
