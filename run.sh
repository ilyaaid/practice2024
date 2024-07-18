#!/bin/bash

rm -rf result
go build -o bin/main cmd/main/main.go 
go build -o bin/main_mpi cmd/main_mpi/main_mpi.go
./bin/main --algo $1 --file test_graphs/$2 --proc-num $3


