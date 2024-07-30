package mympi

func SendTag(comm *Communicator, toID int, tag int) {
	comm.SendBytes(make([]byte, 1), toID, tag)
}

func RecvTag(comm *Communicator, fromID int, tag int) Status {
	_, status := comm.RecvBytes(fromID, tag)
	return status
}

func BcastSendBytes(mes []byte, fromID int, tag int) {
	for i := 0; i < Size(); i++ {
		if i != fromID {
			WorldCommunicator().SendBytes(mes, i, tag)
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
