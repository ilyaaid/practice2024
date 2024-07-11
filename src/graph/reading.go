package graph

import (
	"encoding/csv"
	"log"
	"os"
	"path/filepath"
)

func ReadFromFile(filename string) *Graph {
	file, err := os.Open(filename)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	fileext := filepath.Ext(filename)

	switch fileext {
	case ".csv":
		return readFromCsvFile(file)
	default:
		log.Fatal("unknown file extension (file with graph:" + filename + ")")
		return nil
	}
}

func readFromCsvFile(file *os.File) *Graph {
	csvReader := csv.NewReader(file)

	records, err := csvReader.ReadAll()
	if err != nil {
		log.Fatal(err)
	}

	if csvReader.FieldsPerRecord != 2 {
		log.Fatal("the file \"" + file.Name() + "\" should consist of 2 columns")
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
	for _, row := range records {
		fromInd := getIndex(row[0])
		toInd := getIndex(row[1])
		edges = append(edges, Edge{fromInd, toInd})
	}
	return &Graph{CntVertex: indexCount, Edges: edges, Index2str: index2str}
}
