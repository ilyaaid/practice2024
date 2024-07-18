package graph

type IndexType uint32

type Graph struct {
	VertexCnt IndexType
	Edges     []Edge
	Index2str map[IndexType]string

	CC map[IndexType]IndexType // после выполнения алгоритма сюда кладется результат
}
