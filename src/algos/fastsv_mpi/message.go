package fastsv_mpi

import "CC/graph"

type MessageMPI struct {
	V          graph.IndexType // вершина для которой предлагается родитель
	PPNonConst graph.IndexType // меняющаяся вершина, которая в конце цепочки будет содержать предлагаемого родителя для v
	StartProc  int             // стартовый процесс
}
