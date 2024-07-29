package fastsv_mpi

import (
	"CC/algos/algo_config"
	"CC/algos/algo_types"
	"fmt"
	"log"
	"os"
	"os/exec"
)

func Adapter(conf *algo_config.AlgoConfig) (error) {
	confStr, err := conf.ObjToStr()
	if err != nil {
		return err
	}

	cmd := exec.Command(
		"mpirun",
		"-n", fmt.Sprintf("%d", conf.ProcNum + 1), // +1 так, как добавляется ведущий процесс (master)
		"-oversubscribe",
		"bin/main_mpi",
		"-algo", algo_types.ALGO_fastsv_mpi,
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
		return fmt.Errorf("error in algo %s:\n%s", algo_types.ALGO_fastsv_mpi, err)
	}
	return nil
}
