package fastsv_mpi

import (
	"CC/graph"
	"CC/mympi"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"sync"

	"golang.org/x/sync/errgroup"
)

type StepGoRoutManager struct {
	mutex sync.Mutex

	errG *errgroup.Group
	ctx  context.Context
}

func (stepMan *StepGoRoutManager) Init() {
	stepMan.mutex = sync.Mutex{}
	stepMan.errG, stepMan.ctx = errgroup.WithContext(context.Background())
}

type StepType uint8

const (
	STEP_STOCH_H StepType = iota
	STEP_AGGR_H
	STEP_SHORTCUT_H
)

type Step struct {
	slave   *Slave
	t       StepType
	manager StepGoRoutManager

	chainsCnt int
}

func (step *Step) Init() {
	step.manager.Init()
}

func (step *Step) updateCC(v graph.IndexType, pp graph.IndexType) {
	step.manager.mutex.Lock()
	if step.slave.algo.conf.Variant == VARIANT_DISC_F_PAR {
		if step.slave.fnext[v] > pp {
			step.slave.fnext[v] = pp
			step.slave.fchanged = true
		}
	} else {
		if step.slave.f[v] > pp {
			step.slave.f[v] = pp
			step.slave.fchanged = true
		}
	}
	step.manager.mutex.Unlock()
}

func (step *Step) sendMessageMPI(mes *MessageMPI, toProc int, tag int) error {
	mesBytes, err := json.Marshal(mes)
	if err != nil {
		return err
	}
	step.manager.mutex.Lock()
	step.slave.comm.SendBytes(mesBytes, toProc, tag)

	step.slave.algo.logger.sendTagIntoStep(tag)
	step.manager.mutex.Unlock()
	return nil
}

func (step *Step) getNextTagStep(tag int) int {
	switch tag {
	case TAG_STEP_0:
		return TAG_STEP_1
	case TAG_STEP_1:
		return TAG_STEP_2
	case TAG_STEP_2:
		return TAG_STEP_3
	}
	return TAG_STEP_3
}

func (step *Step) reduceChains() {
	step.manager.mutex.Lock()
	step.chainsCnt = step.chainsCnt - 1
	if step.chainsCnt == 0 {
		mympi.SendTag(step.slave.comm, MASTER_RANK, TAG_END_STEP)
	}
	step.manager.mutex.Unlock()
}

func (step *Step) recvH(mes *MessageMPI, tag int) error {
	switch step.t {
	case STEP_STOCH_H:
		return step.recvStoch(mes, tag)
	case STEP_AGGR_H:
		return step.recvAggr(mes, tag)
	case STEP_SHORTCUT_H:
		return step.recvShortcut(mes, tag)
	}
	return nil
}

func (step *Step) receiving() error {
	for {
		var mes []byte
		var status mympi.Status

		is_exist := false
		for {
			step.manager.mutex.Lock()
			is_exist, _ = step.slave.comm.Iprobe(mympi.AnySource, mympi.AnyTag)
			if is_exist {
				mes, status = step.slave.comm.RecvBytes(mympi.AnySource, mympi.AnyTag)
				step.manager.mutex.Unlock()
				break
			}
			step.manager.mutex.Unlock()
		}

		if err := status.GetError(); err != 0 {
			return fmt.Errorf("step.receving status error:%v", err)
		}

		tag := status.GetTag()
		if tag == TAG_END_STEP {
			break
		}

		step.manager.mutex.Lock()
		step.slave.algo.logger.recvTagIntoStep(tag)
		step.manager.mutex.Unlock()

		var mesObj MessageMPI

		err := json.Unmarshal(mes, &mesObj)
		if err != nil {
			return fmt.Errorf("step.receiving:%v", err)
		}

		err = step.recvH(&mesObj, tag)
		if err != nil {
			return fmt.Errorf("step.receiving^ %v", err)
		}
	}
	return nil
}

func (step *Step) sendH(mes *MessageMPI, tag int) error {
	switch step.t {
	case STEP_STOCH_H:
		return step.sendStoch(mes, tag)
	case STEP_AGGR_H:
		return step.sendAggr(mes, tag)
	case STEP_SHORTCUT_H:
		return step.sendShortcut(mes, tag)
	}
	return nil
}

func (step *Step) getStartMesssageMPI(edge graph.Edge) MessageMPI {
	step.manager.mutex.Lock()

	var mes MessageMPI
	switch step.t {
	case STEP_STOCH_H:
		mes = MessageMPI{V: step.slave.f[edge.V1], PPNonConst: edge.V2, StartProc: step.slave.rank}
	case STEP_AGGR_H:
		mes = MessageMPI{V: edge.V1, PPNonConst: edge.V2}
	case STEP_SHORTCUT_H:
		mes = MessageMPI{V: edge.V1, PPNonConst: step.slave.f[edge.V1]}
	default:
		mes = MessageMPI{}
	}
	step.manager.mutex.Unlock()
	return mes
}

func (step *Step) sending() error {
	if step.t == STEP_SHORTCUT_H {
		for u := range step.slave.f {
			mes := step.getStartMesssageMPI(graph.Edge{V1: u, V2: u})
			err := step.sendH(&mes, TAG_STEP_0)
			if err != nil {
				return fmt.Errorf("sending step.type=%d: %v", step.t, err)
			}
		}
	} else {
		for _, edge := range step.slave.edges {
			var mes MessageMPI
			var err error
			mes = step.getStartMesssageMPI(edge)
			err = step.sendH(&mes, TAG_STEP_0)
			if err != nil {
				return fmt.Errorf("sending step.type=%d: %v", step.t, err)
			}

			if step.slave.IsOwnerOfVertex(edge.V2) {
				mes = step.getStartMesssageMPI(graph.Edge{V1: edge.V2, V2: edge.V1})
				err = step.sendH(&mes, TAG_STEP_0)
				if err != nil {
					return fmt.Errorf("sending step.type=%d: %v", step.t, err)
				}
			}
		}
	}
	return nil
}

func (step *Step) edgeChainsCnt() int {
	cnt := 0
	for _, edge := range step.slave.edges {
		cnt++
		if step.slave.IsOwnerOfVertex(edge.V2) {
			cnt++
		}
	}
	return cnt
}

func (step *Step) getChainsCnt() int {
	switch step.t {
	case STEP_STOCH_H:
		return step.edgeChainsCnt()
	case STEP_AGGR_H:
		return step.edgeChainsCnt()
	case STEP_SHORTCUT_H:
		return len(step.slave.f)
	default:
		return 0
	}
}

func (step *Step) run(t StepType) error {
	step.t = t

	step.manager.errG.SetLimit(2)

	step.chainsCnt = step.getChainsCnt()

	step.slave.slavesComm.Barrier()

	if step.chainsCnt == 0 {
		mympi.SendTag(step.slave.comm, MASTER_RANK, TAG_END_STEP)
		_ = mympi.RecvTag(step.slave.comm, MASTER_RANK, TAG_END_STEP)
	} else {
		step.manager.errG.Go(func() error {
			return step.sending()
		})
		step.manager.errG.Go(func() error {
			return step.receiving()
		})
	}

	if err := step.manager.errG.Wait(); err != nil {
		log.Panicln("runStep wait:", err)
		return err
	}
	return nil
}
