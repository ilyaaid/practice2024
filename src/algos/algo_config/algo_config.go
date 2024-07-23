package algo_config

import (
	"CC/graph"
	"encoding/json"
)

// при изменении или добавлении полей обязательно менять метод UnmarshalJSON
type AlgoConfig struct {
	GrIO graph.GraphIO
	ProcNum int
}

func (conf *AlgoConfig) UnmarshalJSON(data []byte) error {
    type AlgoConfigJSON struct {
		GrIO json.RawMessage
        ProcNum int
    }

    var confj AlgoConfigJSON
    err := json.Unmarshal(data, &confj)
    if err != nil {
        return err
    }
	conf.ProcNum = confj.ProcNum

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
		return nil, err
	}
	return &obj, nil
}
