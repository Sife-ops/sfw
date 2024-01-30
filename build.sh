#!/bin/bash

# todo real makefile

pushd ./cubiomes/
make # todo fix cubiomes makefile
make g0 && cp a.out ../bin/cubiomes
popd

go build -o ./bin/generator \
	./cmd/generator/main.go \
	./cmd/generator/db.go \
	./cmd/generator/container.go \
	./cmd/generator/seed.go \
	./cmd/generator/worldgen.go \
	./cmd/generator/cubiomes.go