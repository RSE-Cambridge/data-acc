#!/bin/bash
set -eux

docker-compose down -v

./build.sh

docker-compose up -d

sleep 10
./register_cluster.sh

docker exec slurmctld bash -c '/usr/local/bin/data-acc/tools/slurm-test.sh'

sleep 10
docker exec slurmctld bash -c "scontrol show burstbuffer"
docker exec slurmctld bash -c "squeue"
