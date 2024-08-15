package ilogger


type ILogger interface {
	Init() error
	Start() error
	Finish() error
	Close() error
}

