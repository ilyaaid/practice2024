package fastsv_mpi

const (
	// для стадии распределения ребер
	TAG_SEND_EDGE int = iota

	//===========================
	// Для алгоритма CC
	TAG_IS_CHANGED 
	TAG_IS_NOT_CHANGED
	TAG_CONTINUE_CC
	TAG_END_STEP // тег для завершения одной итерации в какмо-либо шаге быстрого алгоритма
	
	// для стохастического перевешивания
	TAG_STEP_0
	TAG_STEP_1
	TAG_STEP_2
	TAG_STEP_3

	//============================
	// для вывода результата
	TAG_SEND_RESULT
	TAG_SEND_RESULT_PATH

	TAG_NEXT_PHASE
)
