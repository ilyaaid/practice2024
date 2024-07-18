package graph

type GraphIO interface {
	Read() (*Graph, error)
	WriteCC(map[IndexType]IndexType) error
}
