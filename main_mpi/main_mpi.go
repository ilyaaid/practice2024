package main

import (
	"CC/src/flag_handler"
	"fmt"
	"github.com/sbromberger/gompi"
)


func main() {
	mpi.Start(false)
	defer mpi.Stop()
	
	fmt.Println(mpi.WorldRank(), mpi.WorldSize())

	
	var fh flag_handler.FlagHadler
	fh.Parse()
	
	
}
