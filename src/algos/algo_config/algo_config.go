package algo_config

import (
	"encoding/json"
	"log"
)

type AlgoConfig struct {
	GraphFilename string
	ProcNum int
}

func (conf *AlgoConfig) ObjToStr() string {
	jsonStr, err := json.Marshal(conf)
	if (err != nil) {
		log.Panicln(err)
	}

	return string(jsonStr)
}

func StrToObj(str string) *AlgoConfig {
	var obj AlgoConfig
	err := json.Unmarshal([]byte(str), &obj)
	if (err != nil) {
		log.Panicln(err)
	}
	return &obj
}
