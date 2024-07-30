package fastsv_mpi

func (step *Step) recvShortcut(mes *MessageMPI, tag int) error {
	switch tag {
	case TAG_STEP_0:
		step.manager.mutex.Lock()
		mes.PPNonConst = step.slave.cc[mes.PPNonConst]
		step.manager.mutex.Unlock()

		tag = step.getNextTagStep(tag)
		step.sendShortcut(mes, tag)
	case TAG_STEP_1:
		step.updateCC(mes.V, mes.PPNonConst)

		step.reduceChains()
	}
	return nil
}

func (step *Step) sendShortcut(mes *MessageMPI, tag int) error {
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
		toProc := Vertex2Proc(step.slave.conf, mes.V)
		if toProc == step.slave.rank {
			step.updateCC(mes.V, mes.PPNonConst)
			step.reduceChains()
		} else {
			step.sendMessageMPI(mes, toProc, tag)
		}
	}
	return nil
}
