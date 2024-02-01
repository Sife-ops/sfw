#!/bin/bash

# todo real makefile

pushd ./cubiomes/
make # todo fix cubiomes makefile
make g0 && cp a.out ../bin/cubiomes
popd

go build -o ./bin/generator \
	./cmd/generator/container.go \
	./cmd/generator/cubiomes.go \
	./cmd/generator/main.go \
	./cmd/generator/worldgen.go

go build -o ./bin/scheduler \
	./cmd/scheduler/main.go

go build -o ./bin/cubiomes-worker \
	./cmd/cubiomes-worker/main.go