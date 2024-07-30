package adapter

import (
	"CC/algos/algo_config"
	"log"
)

func Adapter(algo string, conf *algo_config.AlgoConfig) (error) {
	g, err := conf.GrIO.Read()
	if err != nil {
		return err
	}
	ccSearchFunc, err := GetCCSearchFunc(algo)
	if err != nil {
		return err
	}
	ccSearchFunc(g)
	log.Println(g.CC)
	return nil
}
