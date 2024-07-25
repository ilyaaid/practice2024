#!/bin/bash

cd src
go build -o bin/main cmd/main/main.go 
go build -o bin/main_mpi cmd/main_mpi/main_mpi.go
/usr/bin/mpirun -n 4 -oversubscribe bin/main_mpi -algo fastsv_mpi -conf '{"GrIO":{"Filename":"../test_graphs/big-1e2.csv"},"ProcNum":3}'
cd ..
