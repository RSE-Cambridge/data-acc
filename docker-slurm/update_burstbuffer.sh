#!/bin/bash
set -e

git push
docker exec slurmctld bash -c "cd /usr/local/src/burstbuffer && . .venv/bin/activate && git pull && pip install -Ue . && fakewarp help"
docker-compose restart slurmctld

sleep 2

docker exec slurmctld bash -c "cd /data && su slurm -c 'srun --bb=\"capacity=1G\" bash -c \"set\"'"

#
# Use this one to see logs as the job executes
#
# sleep 2
# docker exec slurmctld bash -c "cd /data && su slurm -c 'sbatch -n2 --bbf=buffer.txt --wrap=hostname'"
# docker logs slurmctld -f
