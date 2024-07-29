package mympi

func SendMes(comm *Communicator, mes []byte, toID int, tag int) {
	comm.SendBytes(mes, toID, tag)
}

func RecvMes(comm *Communicator, fromID int, tag int) ([]byte, Status) {
	mes, status := comm.RecvBytes(fromID, tag)
	return mes, status
}

func SendTag(comm *Communicator, toID int, tag int) {
	SendMes(comm, make([]byte, 1), toID, tag)
}

func RecvTag(comm *Communicator, fromID int, tag int) Status {
	_, status := RecvMes(comm, fromID, tag)
	return status
}

func BcastSend(mes []byte, fromID int, tag int) {
	for i := 0; i < Size(); i++ {
		if i != fromID {
			SendMes(WorldCommunicator(), mes, i, tag)
		}
	}
}

func BcastSendTag(fromID, tag int) {
	for i := 0; i < Size(); i++ {
		if i != fromID {
			SendTag(WorldCommunicator(), i, tag)
		}
	}
}
