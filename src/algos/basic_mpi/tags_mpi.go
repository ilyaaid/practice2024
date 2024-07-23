package basic_mpi

const (
	// для стадии распределения ребер
	TAG_SEND_EDGE = iota

	// Для алгоритма CC
	TAG_IS_CHANGED
	TAG_IS_NOT_CHANGED
	TAG_CONTINUE_CC
	
	TAG_SEND_PP

	TAG_SEND_RESULT
	TAG_SEND_RESULT_PATH

	TAG_NEXT_PHASE
)
