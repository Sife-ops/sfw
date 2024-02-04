#!/bin/bash

# todo real makefile

pushd ./cubiomes/
make # todo fix cubiomes makefile
make g0 && cp a.out ../bin/cubiomes
popd

make all