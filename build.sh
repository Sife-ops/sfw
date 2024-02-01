#!/bin/bash

# todo real makefile

pushd ./cubiomes/
make # todo fix cubiomes makefile
make g0 && cp a.out ../bin/cubiomes
popd

go build -o ./bin/generator \
	./cmd/generator/main.go

go build -o ./bin/scheduler \
	./cmd/scheduler/main.go

go build -o ./bin/cubiomes-worker \
	./cmd/cubiomes-worker/main.go

go build -o ./bin/worldgen-worker \
	./cmd/worldgen-worker/main.go