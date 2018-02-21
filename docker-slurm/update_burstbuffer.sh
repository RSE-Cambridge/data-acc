#!/bin/bash
set -x

git push
docker build -t slurm-docker-cluster:17.02.9 . --build-arg BURSTBUFFER_BRANCH=`git show HEAD -q | head -1 | cut -c 8-`

#docker exec slurmctld bash -c "cd /usr/local/src/burstbuffer && . .venv/bin/activate && git remote update && git checkout etcd && git pull && pip install -Ue . && fakewarp help"

docker-compose up -d

sleep 2

docker-compose logs -f &

sleep 5

docker exec slurmctld bash -c "cd /data && su slurm -c 'srun --bb=\"capacity=1G\" bash -c \"set\"'"

sleep 15

#
# Use this one to see logs as the job executes
#
# sleep 2
# docker exec slurmctld bash -c "cd /data && su slurm -c 'sbatch -n2 --bbf=buffer.txt --wrap=hostname'"
# docker logs slurmctld -f

kill %1
