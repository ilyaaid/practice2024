#!/bin/bash

go build -o bin/main main/main.go 
go build -o bin/main_mpi main_mpi/main_mpi.go
./bin/main --algo $1 --file test_graphs/$2 --proc $3


