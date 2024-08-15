package fastsv_mpi

import (
	"fmt"
)

func (step *Step) recvAggr(mes *MessageMPI, tag int) error {
	switch tag {
	case TAG_STEP_0, TAG_STEP_1:
		step.manager.mutex.Lock()
		mes.PPNonConst = step.slave.f[mes.PPNonConst]
		step.manager.mutex.Unlock()

		tag = step.getNextTagStep(tag)

		// TODO обработка ошибок
		step.sendAggr(mes, tag)
	case TAG_STEP_2:
		step.updateCC(mes.V, mes.PPNonConst)

		step.reduceChains()
	default:
		return fmt.Errorf("wrong TAG in recvAggr: %d", tag)
	}
	return nil
}

func (step *Step) sendAggr(mes *MessageMPI, tag int) error {
	switch tag {
	case TAG_STEP_0:
		toProc := step.slave.algo.getSlave(mes.PPNonConst)

		if toProc == step.slave.rank {
			step.manager.mutex.Lock()
			mes.PPNonConst = step.slave.f[mes.PPNonConst]
			step.manager.mutex.Unlock()

			tag = step.getNextTagStep(tag)
		} else {
			// TODO обработка ошибок добавить
			step.sendMessageMPI(mes, toProc, tag)

			return nil
		}
		fallthrough
	case TAG_STEP_1:
		toProc := step.slave.algo.getSlave(mes.PPNonConst)

		if toProc == step.slave.rank {
			step.manager.mutex.Lock()
			mes.PPNonConst = step.slave.f[mes.PPNonConst]
			step.manager.mutex.Unlock()

			tag = step.getNextTagStep(tag)
		} else {
			// TODO обработка ошибок добавить
			step.sendMessageMPI(mes, toProc, tag)

			return nil
		}
		fallthrough
	case TAG_STEP_2:
		toProc := step.slave.algo.getSlave(mes.V)

		if toProc == step.slave.rank {
			step.updateCC(mes.V, mes.PPNonConst)
			step.reduceChains()
		} else {
			step.sendMessageMPI(mes, toProc, tag)
		}
	}
	return nil
}
