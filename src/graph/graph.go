package graph

import (
	"encoding/json"
	"log"
)

type IndexType uint32

type Edge struct {
	V1 IndexType
	V2 IndexType
}

func (edge *Edge) ToStr() string {
	bytes, err := json.Marshal(edge)
	if (err != nil) {
		log.Panicln(err)
	}
	return string(bytes)
}

func StrToEdgeObj(str string) *Edge {
	var edge Edge
	err := json.Unmarshal([]byte(str), &edge)
	if (err != nil) {
		log.Panicln(err)
	}
	return &edge
}

type Graph struct {
	VertexCnt IndexType
	Edges     []Edge
	Index2str map[IndexType]string

	CC map[IndexType]IndexType // после выполнения алгоритма сюда кладется результат
}
