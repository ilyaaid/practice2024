package graph

import "encoding/json"

type Edge struct {
	V1 IndexType
	V2 IndexType
}

func (edge *Edge) ToStr() (string, error) {
	bytes, err := json.Marshal(edge)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

func StrToEdgeObj(str string) (*Edge, error) {
	var edge Edge
	err := json.Unmarshal([]byte(str), &edge)
	if err != nil {
		return nil, err
	}
	return &edge, nil
}
