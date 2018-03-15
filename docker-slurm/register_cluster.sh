#!/bin/bash
set -e

docker exec slurmctld bash -c "/usr/bin/sacctmgr --immediate add cluster name=linux" && \
docker-compose restart slurmdbd slurmctld

echo
echo GlusterFS setup
docker exec gluster1 bash -c "gluster peer probe gluster2"
docker exec gluster1 bash -c "gluster peer probe gluster3"
docker exec gluster1 bash -c "gluster pool list"
docker exec gluster2 bash -c "gluster pool list"
docker exec gluster3 bash -c "gluster pool list"

echo
for i in `seq 1 2`;
do
  docker exec gluster1 bash -c "mkdir -p /data/glusterfs/nvme${i}n1/brick"
done
for i in `seq 1 2`;
do
  docker exec gluster2 bash -c "mkdir -p /data/glusterfs/nvme${i}n1/brick"
done

echo
docker exec gluster1 bash -c "gluster volume create buffer1 gluster1:/data/glusterfs/nvme1n1/brick gluster2:/data/glusterfs/nvme1n1/brick force"
docker exec gluster1 bash -c "gluster volume start buffer1"
echo
docker run --privileged --rm --net dockerslurm_default gluster/glusterfs-client bash -c "mount -t glusterfs gluster1:/buffer1 /mnt && echo 'We have written to a shared file on `date`' >/mnt/test"
docker run --privileged --rm --net dockerslurm_default gluster/glusterfs-client bash -c "mount -t glusterfs gluster1:/buffer1 /mnt && cat /mnt/test"

echo
for i in `seq 1 2`;
do
  docker exec gluster1 bash -c "rm -rf /data/glusterfs/nvme${i}n1/brick"
done
for i in `seq 1 2`;
do
  docker exec gluster2 bash -c "rm -rf /data/glusterfs/nvme${i}n1/brick"
done
docker exec gluster1 bash -c "gluster --mode=script volume stop buffer1 force"
docker exec gluster1 bash -c "gluster --mode=script volume delete buffer1"

echo
SSH_PUB_KEY=`docker exec slurmctld bash -c "cat /root/.ssh/id_rsa.pub"`
docker exec gluster1 bash -c "mkdir -p /root/.ssh && echo $SSH_PUB_KEY > /root/.ssh/authorized_keys"
docker exec slurmctld bash -c "ssh -oStrictHostKeyChecking=no -p 2222 gluster1 hostname"
docker exec gluster2 bash -c "mkdir -p /root/.ssh && echo $SSH_PUB_KEY > /root/.ssh/authorized_keys"
docker exec slurmctld bash -c "ssh -oStrictHostKeyChecking=no -p 2222 gluster2 hostname"
