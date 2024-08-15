#!/bin/bash

rootdir=$(pwd)
cd src
go build -o ${rootdir}/bin/main cmd/main/main.go 
go build -o ${rootdir}/bin/main_mpi cmd/main_mpi/main_mpi.go
cd ..
