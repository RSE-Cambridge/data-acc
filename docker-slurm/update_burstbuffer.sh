#!/bin/bash
set -x

git push
docker build -t slurm-docker-cluster:17.02.9 . --build-arg BURSTBUFFER_BRANCH=`git show HEAD -q | head -1 | cut -c 8-`

#docker exec slurmctld bash -c "cd /usr/local/src/burstbuffer && . .venv/bin/activate && git remote update && git checkout etcd && git pull && pip install -Ue . && fakewarp help"

docker-compose up -d

sleep 2

docker exec slurmctld bash -c "cd /data && su slurm -c 'srun --bb=\"capacity=1G\" bash -c \"set\"'"

#
# Use this one to see logs as the job executes
#
# sleep 2
# docker exec slurmctld bash -c "cd /data && su slurm -c 'sbatch -n2 --bbf=buffer.txt --wrap=hostname'"
# docker logs slurmctld -f

docker-compose logs -f &

sleep 5

echo Assign some burst buffers

docker exec slurmctld bash -c "ETCDCTL_API=3 etcdctl --endpoints=http://etcdproxy1:2379 del --prefix bufferhosts/assigned_slices"
docker exec slurmctld bash -c "ETCDCTL_API=3 etcdctl --endpoints=http://etcdproxy1:2379 put bufferhosts/assigned_slices/fakenode1/nvme0n1 buffer/fakebuffer1"
docker exec slurmctld bash -c "ETCDCTL_API=3 etcdctl --endpoints=http://etcdproxy1:2379 put bufferhosts/assigned_slices/fakenode1/nvme1n1 buffer/fakebuffer1"
docker exec slurmctld bash -c "ETCDCTL_API=3 etcdctl --endpoints=http://etcdproxy1:2379 put bufferhosts/assigned_slices/fakenode2/nvme9n1 buffer/fakebuffer1"
docker exec slurmctld bash -c "ETCDCTL_API=3 etcdctl --endpoints=http://etcdproxy1:2379 put bufferhosts/assigned_slices/fakenode3/nvme9n1 buffer/fakebuffer1"
docker exec slurmctld bash -c "ETCDCTL_API=3 etcdctl --endpoints=http://etcdproxy1:2379 put bufferhosts/assigned_slices/fakenode1/nvme3n1 buffer/fakebuffer2"

sleep 5

echo Notice how they are picked up, now we delete them...

docker exec slurmctld bash -c "ETCDCTL_API=3 etcdctl --endpoints=http://etcdproxy1:2379 del --prefix bufferhosts/assigned_slices"

sleep 5

kill %1
