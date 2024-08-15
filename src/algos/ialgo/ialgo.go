package ialgo

import (
	"CC/algos/algo_config"
	"CC/algos/ilogger"
)

type IAlgo interface {
	Init(*algo_config.AlgoConfig) error
	GetLogger() ilogger.ILogger
	Run() error
	Close() error
}
