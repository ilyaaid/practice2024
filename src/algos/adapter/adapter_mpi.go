package adapter

import (
	"CC/algos/algo_config"
	"fmt"
	"log"
	"os"
	"os/exec"
)

func AdapterMPI(algo string, conf *algo_config.AlgoConfig) (error) {
	confStr, err := conf.ObjToStr()
	if err != nil {
		return err
	}

	cmd := exec.Command(
		"mpirun",
		"-np", fmt.Sprintf("%d", conf.ProcNum + 1), // +1 так, как добавляется ведущий процесс (master)
		"-oversubscribe",
		"bin/main_mpi",
		"-algo", algo,
		"-conf", confStr,
	)
	log.Println("COMMAND FOR MPI: \n", cmd)

	// перенаправляем вывод
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("error in algo %s:\n%s", algo, err)
	}

	if cmd.Process != nil {
		cmd.Process.Kill()
	}

	return nil
}
