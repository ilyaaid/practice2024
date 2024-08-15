package graph

type GraphIO interface {
	Read() (*Graph, error)
	Write(map[IndexType]IndexType) error
}
