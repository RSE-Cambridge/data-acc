#!/bin/bash
set -e

docker exec slurmctld bash -c "/usr/bin/sacctmgr --immediate add cluster name=linux" && \
docker-compose restart slurmdbd slurmctld

echo
echo GlusterFS setup
docker exec gluster1 bash -c "gluster peer probe gluster2"
docker exec gluster1 bash -c "gluster pool list"
docker exec gluster2 bash -c "gluster pool list"

echo
for i in `seq 1 5`;
do
  docker exec gluster1 bash -c "mkdir -p /data/glusterfs/nvme$i/brick"
done
for i in `seq 1 5`;
do
  docker exec gluster2 bash -c "mkdir -p /data/glusterfs/nvme$i/brick"
done

echo
docker exec gluster1 bash -c "gluster volume create buffer1 gluster1:/data/glusterfs/nvme1/brick gluster2:/data/glusterfs/fakebrick/nvme1/brick force" || true
docker exec gluster1 bash -c "gluster volume start buffer1" || true
echo
docker run --privileged --rm --net dockerslurm_default gluster/glusterfs-client bash -c "mount -t glusterfs gluster1:/buffer1 /mnt && echo 'We have written to a shared file on `date`' >/mnt/test"
docker run --privileged --rm --net dockerslurm_default gluster/glusterfs-client bash -c "mount -t glusterfs gluster1:/buffer1 /mnt && cat /mnt/test"
