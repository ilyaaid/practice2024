package basic_mpi

import (
	"CC/src/algos/algo_config"
	"CC/src/algos/algo_types"
	"CC/src/graph"
	"fmt"
	"log"
	"os"
	"os/exec"
)

func Adapter(conf *algo_config.AlgoConfig) (*graph.Graph, error) {
	confStr, err := conf.ObjToStr()
	if err != nil {
		return nil, err
	}

	cmd := exec.Command(
		"mpirun",
		"-n", fmt.Sprintf("%d", conf.ProcNum + 1), // +1 так, как добавляется ведущий процесс (master)
		"-oversubscribe",
		"bin/main_mpi",
		"-algo", algo_types.ALGO_basic_mpi,
		"-conf", confStr,
	)
	log.Println("COMMAND FOR MPI: \n", cmd)

	// перенаправляем вывод
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	cmd.Start()
	defer cmd.Process.Kill()

	// ждем, пока mpi отработает и выводим, что он написал
	if err := cmd.Wait(); err != nil {
		return nil, fmt.Errorf("error in algo %s:\n%s", algo_types.ALGO_basic_mpi, err)
	}
	return &graph.Graph{}, nil
}
