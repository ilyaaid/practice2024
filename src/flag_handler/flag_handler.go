package flag_handler

import (
	"CC/src/algos/algo_types"
	"flag"
	"fmt"
	"io"
	"log"
)

// все возможные флаги командной строки при вызове программы
const (
	FLAG_algo = "algo"
	FLAG_file = "file"
	FLAG_proc_num = "proc-num"
	FLAG_conf = "conf"
)

type FlagHadler struct {
	Algo string
	File string
	Proc int
	Conf string
}

func (fh *FlagHadler) Parse() {
	flag.StringVar(&fh.Algo, FLAG_algo, "", 
		"algorithm type (" + algo_types.ALGO_basic + ", " + algo_types.ALGO_basic_mpi +", ...)")
	flag.StringVar(&fh.File, FLAG_file, "", "file with graph")
	flag.IntVar(&fh.Proc, FLAG_proc_num, 0, "number of processes (for MPI)")
	flag.StringVar(&fh.Conf, FLAG_conf, "", "config json string (for MPI)")

	flag.Parse()

}

func (fh *FlagHadler) Print(w io.Writer) {
	str := ""
	flag.VisitAll(func(f *flag.Flag) {
		str += fmt.Sprintf("<FLAG>    %s: %s\n", f.Name, f.Value)
	})
	log.Println(str)
	// TODO w.Write([]byte(str))
}
