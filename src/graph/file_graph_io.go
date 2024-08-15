package graph

import (
	"encoding/csv"
	"fmt"
	"os"
	"path/filepath"
)

type FileGraphIO struct {
	Filename string
}

func (fGrIO *FileGraphIO) Read() (*Graph, error) {
	file, err := os.Open(fGrIO.Filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	fileext := filepath.Ext(fGrIO.Filename)

	switch fileext {
	case ".csv":
		return readFromCsvFile(file)
	default:
		return nil, fmt.Errorf("unknown file extension (file with graph: %s)", fGrIO.Filename)
	}
}

func readFromCsvFile(file *os.File) (*Graph, error) {
	csvReader := csv.NewReader(file)

	records, err := csvReader.ReadAll()
	if err != nil {
		return nil, err
	}

	if csvReader.FieldsPerRecord != 2 {
		return nil, fmt.Errorf("the file \"%s\" should consist of 2 columns", file.Name())
	}

	indexCount := IndexType(0)
	str2index := make(map[string]IndexType)
	index2str := make(map[IndexType]string)
	edges := make([]Edge, 0)

	getIndex := func(str string) IndexType {
		ind, ok := str2index[str]
		if !ok {
			str2index[str] = indexCount
			index2str[indexCount] = str
			ind = indexCount
			indexCount++
		}
		return ind
	}
	
	cc := make(map[IndexType]IndexType)
	for _, row := range records {
		v1Ind := getIndex(row[0])
		v2Ind := getIndex(row[1])

		edge := Edge{V1: v1Ind, V2:v2Ind}
		edges = append(edges, edge)
		cc[edge.V1] = edge.V1
		cc[edge.V2] = edge.V2
	}
	return &Graph{ 
		Edges: edges, 
		Index2str: index2str, 
		CC: cc}, nil
}


func (fGrIO *FileGraphIO) Write(cc map[IndexType]IndexType) error {
	return nil
}
