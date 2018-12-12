#!/bin/bash
set -eux

cd ../
make clean
make tar

cd docker-slurm
rm -rf ./bin
mkdir ./bin
cp ../bin/*.tgz ./bin
mv `ls bin/*.tgz` ./bin/data-acc.tgz

docker-compose build
#docker-compose push
