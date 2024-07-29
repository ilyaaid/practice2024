package basic

import (
	"CC/algos/algo_config"
	"log"
)

func Adapter(conf *algo_config.AlgoConfig) (error) {
	g, err := conf.GrIO.Read()
	if err != nil {
		return err
	}
	CCSearch(g)
	log.Println(g.CC)
	return nil
}
