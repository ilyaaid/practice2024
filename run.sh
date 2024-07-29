#!/bin/bash

rootdir=$(pwd)
cd src
go build -o ${rootdir}/bin/main cmd/main/main.go 
go build -o ${rootdir}/bin/main_mpi cmd/main_mpi/main_mpi.go
cd ..
./bin/main --algo $1 --file test_graphs/$2 --proc-num $3

