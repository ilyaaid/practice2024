package algo_config

import (
	"CC/graph"
	"encoding/json"
	"log"
)

// при изменении или добавлении полей обязательно менять метод UnmarshalJSON
type AlgoConfig struct {
	GrIO graph.GraphIO
	ProcNum int
	Variant string
}

func (conf *AlgoConfig) UnmarshalJSON(data []byte) error {
    type AlgoConfigJSON struct {
		GrIO json.RawMessage
        ProcNum int
		Variant string
    }

    var confj AlgoConfigJSON
    err := json.Unmarshal(data, &confj)
    if err != nil {
        return err
    }
	conf.ProcNum = confj.ProcNum
	conf.Variant = confj.Variant

	var v1 graph.FileGraphIO
	err = json.Unmarshal(confj.GrIO, &v1)
	if err == nil {
		conf.GrIO = &v1
		return nil
	}
    return err
}

func (conf *AlgoConfig) ObjToStr() (string, error) {
	jsonStr, err := json.Marshal(conf)
	if (err != nil) {
		return "", err
	}

	return string(jsonStr), nil
}

func StrToObj(str string) (*AlgoConfig, error) {
	var obj AlgoConfig
	err := json.Unmarshal([]byte(str), &obj)
	if (err != nil) {
		log.Println("conf strtoobj: problem with unmarshal", err)
		return nil, err
	}
	return &obj, nil
}
