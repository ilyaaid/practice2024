package graph

import "encoding/json"

type Edge struct {
	V1 IndexType
	V2 IndexType
}

func (edge *Edge) ToBytes() ([]byte, error) {
	bytes, err := json.Marshal(edge)
	if err != nil {
		return []byte{}, err
	}
	return bytes, nil
}

func StrToEdgeObj(str []byte) (*Edge, error) {
	var edge Edge
	err := json.Unmarshal(str, &edge)
	if err != nil {
		return nil, err
	}
	return &edge, nil
}
