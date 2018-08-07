#!/bin/bash

set -eux

make

scp ./bin/amd64/data-acc-brick-host centos@128.232.224.186:~/

cd docker-slurm

rm -rf ./bin
mkdir -p ./bin
cp -r ../bin/amd64/* ./bin

docker-compose build
docker-compose push
