package fastsv_mpi

import (
	"fmt"
)

func (step *Step) recvStoch(mes *MessageMPI, tag int) error {
	var err error
	switch tag {
	case TAG_STEP_0:
		step.manager.mutex.Lock()
		mes.PPNonConst = step.slave.cc[mes.PPNonConst]
		step.manager.mutex.Unlock()

		tag = step.getNextTagStep(tag)
		err = step.sendStoch(mes, tag)
	case TAG_STEP_1:
		step.manager.mutex.Lock()
		mes.PPNonConst = step.slave.cc[mes.PPNonConst]
		step.manager.mutex.Unlock()

		tag = step.getNextTagStep(tag)
		err = step.sendStoch(mes, tag)
	case TAG_STEP_2:
		step.updateCC(mes.V, mes.PPNonConst)

		tag = step.getNextTagStep(tag)
		err = step.sendStoch(mes, tag)
	case TAG_STEP_3:
		step.reduceChains()
	default:
		err = fmt.Errorf("wrong TAG: %d", tag)
	}
	if err != nil {
		return fmt.Errorf("step.recvStoch: %v", err)
	}
	return nil
}


func (step *Step) sendStoch(mes *MessageMPI, tag int) error {
	switch tag {
	case TAG_STEP_0:
		toProc := Vertex2Proc(step.slave.conf, mes.PPNonConst)

		if toProc == step.slave.rank {
			step.manager.mutex.Lock()
			mes.PPNonConst = step.slave.cc[mes.PPNonConst]
			step.manager.mutex.Unlock()

			tag = step.getNextTagStep(tag)
		} else {
			step.sendMessageMPI(mes, toProc, tag)
			return nil
		}
		fallthrough
	case TAG_STEP_1:
		toProc := Vertex2Proc(step.slave.conf, mes.PPNonConst)

		if toProc == step.slave.rank {
			step.manager.mutex.Lock()
			mes.PPNonConst = step.slave.cc[mes.PPNonConst]
			step.manager.mutex.Unlock()

			tag = step.getNextTagStep(tag)
		} else {
			step.sendMessageMPI(mes, toProc, tag)
			return nil
		}
		fallthrough
	case TAG_STEP_2:
		toProc := Vertex2Proc(step.slave.conf, mes.V)
		if toProc == step.slave.rank {
			step.updateCC(mes.V, mes.PPNonConst)
			tag = TAG_STEP_3
		} else {
			step.sendMessageMPI(mes, toProc, tag)
			return nil
		}
		fallthrough
	case TAG_STEP_3: 
		if step.slave.rank == mes.StartProc {
			step.reduceChains()
		} else {
			step.sendMessageMPI(mes, mes.StartProc, tag)
			return nil
		}
	default:
		return fmt.Errorf("wrong TAG in sendStoch: %d", tag)
	}
	return nil
}
