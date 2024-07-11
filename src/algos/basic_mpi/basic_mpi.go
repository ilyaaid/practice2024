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

func Adapter(conf *algo_config.AlgoConfig) *graph.Graph {
	cmd := exec.Command(
		"mpirun",
		"-n", fmt.Sprintf("%d", conf.Proc),
		"-oversubscribe",
		"bin/main_mpi",
		"-algo", algo_types.ALGO_basic_mpi,
		"-conf", conf.ObjToStr(),
	)

	// перенаправляем вывод
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	fmt.Println("=======COMMAND FOR MPI========\n", cmd, "\n=========================")

	cmd.Start()
	defer cmd.Process.Kill()

	// ждем, пока mpi отработает и выводим, что он написал
	if err := cmd.Wait(); err != nil {
		fmt.Println("error in algo " + algo_types.ALGO_basic_mpi)
		log.Panic(err)
	}
	return &graph.Graph{}
}
