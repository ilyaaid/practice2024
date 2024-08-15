package fastsv_mpi

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
	StepType     StepType
	IterationNum int

	SendTag0 int
	SendTag1 int
	SendTag2 int
	SendTag3 int
	RecvTag0 int
	RecvTag1 int
	RecvTag2 int
	RecvTag3 int
}

type Logger struct {
	csvFile *os.File
	records []*record

	algoStartTime time.Time

	sendTagsCnt map[int]int
	recvTagsCnt map[int]int

	stopChan chan struct{}

	step    StepType
	iterNum int

	mutex *sync.Mutex

	isStarted bool
}

func (logger *Logger) Init() error {
	logger.sendTagsCnt = make(map[int]int)
	logger.recvTagsCnt = make(map[int]int)
	logger.mutex = &sync.Mutex{}
	logger.isStarted = false
	return nil
}

func (logger *Logger) Start() error {
	// все логи будут помещаться в новую папку
	// TODO добавить еще время
	dirName := "log/fastsv_" + time.Now().Format("2006-01-02_15:04:05")
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
		commonFile.WriteString(fmt.Sprintf("Algorithm duration: %s", algoDuration))
	}

	return nil
}

func (logger *Logger) Close() error {

	return nil
}

func (logger *Logger) sendTagIntoStep(tag int) {
	if !logger.isStarted {
		return
	}

	logger.mutex.Lock()
	logger.sendTagsCnt[tag]++
	logger.mutex.Unlock()
}

func (logger *Logger) recvTagIntoStep(tag int) {
	if !logger.isStarted {
		return
	}

	logger.mutex.Lock()
	logger.recvTagsCnt[tag]++
	logger.mutex.Unlock()
}

func (logger *Logger) addRecord() {
	if !logger.isStarted {
		return
	}

	logger.mutex.Lock()

	logger.records = append(logger.records, &record{
		Time:         time.Since(logger.algoStartTime),
		StepType:     logger.step,
		IterationNum: logger.iterNum,
		SendTag0:     logger.sendTagsCnt[TAG_STEP_0],
		SendTag1:     logger.sendTagsCnt[TAG_STEP_1],
		SendTag2:     logger.sendTagsCnt[TAG_STEP_2],
		SendTag3:     logger.sendTagsCnt[TAG_STEP_3],
		RecvTag0:     logger.recvTagsCnt[TAG_STEP_0],
		RecvTag1:     logger.recvTagsCnt[TAG_STEP_1],
		RecvTag2:     logger.recvTagsCnt[TAG_STEP_2],
		RecvTag3:     logger.recvTagsCnt[TAG_STEP_3],
	})

	gocsv.MarshalWithoutHeaders(logger.records, logger.csvFile)
	logger.records = make([]*record, 0)

	logger.mutex.Unlock()
}

func (logger *Logger) periodicRecord() {
	if !logger.isStarted {
		return
	}

	ticker := time.NewTicker(1 * time.Second)
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

func (logger *Logger) beginStep(step StepType) {
	if !logger.isStarted {
		return
	}

	logger.mutex.Lock()
	logger.step = step
	logger.mutex.Unlock()

	logger.stopChan = make(chan struct{})
	go logger.periodicRecord()
}

func (logger *Logger) endStep() {
	if !logger.isStarted {
		return
	}

	close(logger.stopChan)
	logger.addRecord()

	logger.mutex.Lock()
	logger.sendTagsCnt = make(map[int]int)
	logger.recvTagsCnt = make(map[int]int)
	logger.mutex.Unlock()
}

func (logger *Logger) beginIterations() {
	if !logger.isStarted {
		return
	}

}

func (logger *Logger) nextIteration() {
	if !logger.isStarted {
		return
	}

	logger.mutex.Lock()
	logger.iterNum++
	logger.mutex.Unlock()
}

func (logger *Logger) endIterations() {
	if !logger.isStarted {
		return
	}

}
