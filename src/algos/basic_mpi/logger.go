package basic_mpi

import (
	"CC/mympi"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/gocarina/gocsv"
)

type record struct {
	Time         time.Duration
	IterationNum int

	SendTag int
	RecvTag int
}

type Logger struct {
	csvFile *os.File
	records []*record

	stopChan chan struct{}

	algoStartTime time.Time

	isStarted bool

	iterNum int
	sendTagCnt int
	recvTagCnt int

	mutex *sync.Mutex
}

func (logger *Logger) Init() error {
	logger.isStarted = false
	logger.mutex = &sync.Mutex{}
	return nil
}

func (logger *Logger) Start() error {
	// все логи будут помещаться в новую папку
	// TODO добавить еще время 
	dirName := "log/basic_" + time.Now().Format("2006-01-02_15:04:05")
	if _, err := os.Stat(dirName); os.IsNotExist(err) {
		os.Mkdir(dirName, 0755)
	}
	csvfilename := fmt.Sprintf(dirName+"/log%d.csv", mympi.WorldRank())

	var err error
	logger.csvFile, err = os.Create(csvfilename)
	if err != nil {
		return fmt.Errorf("AlgoLogger.Init(): Error creating file: %v", err)
	}

	err = gocsv.MarshalFile(logger.records, logger.csvFile)
	if err != nil {
		return fmt.Errorf("AlgoLogger.Init(): Error marshaling CSV: %v", err)
	}

	logger.isStarted = true
	logger.algoStartTime = time.Now()

	return nil
}

func (logger *Logger) Finish() error {
	if !logger.isStarted {
		return nil
	}

	err := logger.csvFile.Close()
	if err != nil {
		return err
	}

	algoDuration := time.Since(logger.algoStartTime)
	mympi.WorldCommunicator().Barrier()
	if mympi.WorldRank() == MASTER_RANK {
		commonFile, _ := os.Create(filepath.Dir(logger.csvFile.Name()) + "/common.txt")
		commonFile.WriteString(fmt.Sprintf("Algorithm duration: %s\n", algoDuration))
		commonFile.WriteString(fmt.Sprintf("Iteration count: %d", logger.iterNum))
	}
	return nil
}

func (logger *Logger) Close() error {

	return nil
}


func (logger *Logger) sendTag() {
	if !logger.isStarted {
		return
	}

	logger.mutex.Lock()
	logger.sendTagCnt++
	logger.mutex.Unlock()
}

func (logger *Logger) recvTag() {
	if !logger.isStarted {
		return
	}

	logger.mutex.Lock()
	logger.recvTagCnt++
	logger.mutex.Unlock()
}

func (logger *Logger) addRecord() {
	if !logger.isStarted {
		return
	}

	logger.mutex.Lock()

	logger.records = append(logger.records, &record{
		Time:         time.Since(logger.algoStartTime),
		IterationNum: logger.iterNum,
		SendTag:     logger.sendTagCnt,
		RecvTag:     logger.recvTagCnt,
	})

	gocsv.MarshalWithoutHeaders(logger.records, logger.csvFile)
	logger.records = make([]*record, 0)

	logger.mutex.Unlock()
}


func (logger *Logger) periodicRecord() {
	if !logger.isStarted {
		return
	}

	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			logger.addRecord()
		case <-logger.stopChan:
			return
		}
	}
}

func (logger *Logger) beginIteration() {
	if !logger.isStarted {
		return
	}
	logger.sendTagCnt = 0
	logger.recvTagCnt = 0

	logger.stopChan = make(chan struct{})
	go logger.periodicRecord()
}

func (logger *Logger) endIteration() {
	if !logger.isStarted {
		return
	}

	close(logger.stopChan)
	logger.addRecord()

	logger.iterNum++
}


func (logger *Logger) beginIterations() {
	if !logger.isStarted {
		return
	}

	
}

func (logger *Logger) endIterations() {
	if !logger.isStarted {
		return
	}
}
