package fastsv_mpi

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"sync"

	mpi "github.com/sbromberger/gompi"
	"golang.org/x/sync/errgroup"
)

func (slave *Slave) receivingStochH(ctx *context.Context, chainsCnt *int, mutex *sync.Mutex) error {
	for {
		log.Println("before recv")
		mesStr, status := slave.comm.RecvBytes(mpi.AnySource, mpi.AnyTag)

		if err := status.GetError(); err != 0 {
			log.Println("recv status error:", err)
			return fmt.Errorf("recv status error:%v", err)
		}

		tag := status.GetTag()
		if tag == TAG_END_STEP {
			log.Println("TAG_END_STEP")
			break
		}

		var mes MessageMPI
		for mesStr[len(mesStr) - 1] == 0 {
			mesStr = mesStr[:len(mesStr) - 1]
		}
		err := json.Unmarshal(mesStr, &mes)
		if err != nil {
			log.Panicln("HELLO!!!!!", status.GetSource(), tag, mesStr, err)
			return fmt.Errorf("receivingStochH:%v", err)
		}
		log.Println("HELLO!!!!!", status.GetSource(), tag, mesStr, err)

		log.Println("recv mes=", mes, "from tag=", tag)

		switch tag {
		case TAG_STEP_0:
			mes.PPNonConst = slave.cc[mes.PPNonConst]
			err := slave.sendStochH(mes, TAG_STEP_1, chainsCnt, mutex)
			if err != nil {
				return fmt.Errorf("send TAG_STEP_1: %v", err)
			}
		case TAG_STEP_1:
			mes.PPNonConst = slave.cc[mes.PPNonConst]
			err := slave.sendStochH(mes, TAG_STEP_2, chainsCnt, mutex)
			if err != nil {
				return fmt.Errorf("send TAG_STEP_2: %v", err)
			}
		case TAG_STEP_2:
			if slave.ccnext[mes.V] > mes.PPNonConst {
				slave.ccnext[mes.V] = mes.PPNonConst
				slave.changed = true
			}
			err := slave.sendStochH(mes, TAG_STEP_3, chainsCnt, mutex)
			if err != nil {
				return fmt.Errorf("send TAG_STEP_3: %v", err)
			}
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
		// if slave.IsOwnerOfVertex(mes.PPNonConst) {
		// 	mes.PPNonConst = slave.cc[mes.PPNonConst]
		// 	tag = TAG_STEP_1
		// } else {
			toProc := Vertex2Proc(slave.conf, mes.PPNonConst)
			mesBytes, err := json.Marshal(mes)
			if err != nil {
				return err
			}
			log.Println(string(mesBytes))
			slave.comm.SendBytes(mesBytes, toProc, tag)
			log.Println("send: HELLO!!!!!", tag, mesBytes)
			break
		// }
		fallthrough
	case TAG_STEP_1:
		// if slave.IsOwnerOfVertex(mes.PPNonConst) {
		// 	mes.PPNonConst = slave.cc[mes.PPNonConst]
		// 	tag = TAG_STEP_2
		// } else {
			toProc := Vertex2Proc(slave.conf, mes.PPNonConst)
			mesBytes, err := json.Marshal(mes)
			if err != nil {
				return err
			}
			slave.comm.SendBytes(mesBytes, toProc, tag)
			log.Println("send: HELLO!!!!!", tag, mesBytes)
			break
		// }
		fallthrough
	case TAG_STEP_2:
		// if slave.IsOwnerOfVertex(mes.V) {
		// 	if slave.ccnext[mes.V] > mes.PPNonConst {
		// 		slave.ccnext[mes.V] = mes.PPNonConst
		// 		slave.changed = true
		// 	}
		// 	tag = TAG_STEP_3
		// } else {
			log.Println("send!!", mes)
			toProc := Vertex2Proc(slave.conf, mes.V)
			mesBytes, err := json.Marshal(mes)
			if err != nil {
				return err
			}
			slave.comm.SendBytes(mesBytes, toProc, tag)
			log.Println("send: HELLO!!!!!", tag, mesBytes)
			break
		// }
		fallthrough
	case TAG_STEP_3:
		// if slave.rank == mes.StartProc {
		// 	mutex.Lock()
		// 	*chainsCnt = *chainsCnt - 1
		// 	if *chainsCnt == 0 {
		// 		sendTag(slave.comm, MASTER_RANK, TAG_END_STEP)
		// 		*chainsCnt = -1
		// 	}
		// 	mutex.Unlock()
		// } else {
			mesBytes, err := json.Marshal(mes)
			if err != nil {
				return err
			}
			slave.comm.SendBytes(mesBytes, mes.StartProc, tag)
		// }
	}
	return nil
}

func (slave *Slave) sendingStochH(ctx *context.Context, chainsCnt *int, mutex *sync.Mutex) error {
	for _, edge := range slave.edges {
		mes := MessageMPI{V: slave.cc[edge.V1], PPNonConst: edge.V2, StartProc: slave.rank}
		err := slave.sendStochH(mes, TAG_STEP_0, chainsCnt, mutex)
		if err != nil {
			return fmt.Errorf("sendingStoch: %v", err)
		}
	}
	return nil
}

func (slave *Slave) runStochH() error {
	g, ctx := errgroup.WithContext(context.Background())
	g.SetLimit(2)
	chainsCnt := len(slave.edges)

	if chainsCnt == 0 {
		sendTag(slave.comm, MASTER_RANK, TAG_END_STEP)
		_ = recvTag(slave.comm, MASTER_RANK, TAG_END_STEP)
	} else {
		mutex := &sync.Mutex{}
		g.Go(func() error {
			return slave.sendingStochH(&ctx, &chainsCnt, mutex)
		})
		g.Go(func() error {
			return slave.receivingStochH(&ctx, &chainsCnt, mutex)
		})
	}

	if err := g.Wait(); err != nil {
		log.Panicln("!!!!!!wait:", err)
		return err
	}
	return nil
}
