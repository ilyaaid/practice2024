package graph

type IndexType uint32

type Edge struct {
	V1 IndexType
	V2 IndexType
}

type Graph struct {
	CntVertex IndexType
	Edges []Edge
	Index2str map[IndexType]string

	CC map[IndexType]IndexType
}

