package algo_config

import (
	"encoding/json"
	"log"
)

type AlgoConfig struct {
	GraphFilename string
	Proc uint
}

func (conf *AlgoConfig) ObjToStr() string {
	jsonStr, err := json.Marshal(conf)
	if (err != nil) {
		log.Panicln(err)
	}

	return string(jsonStr)
}
