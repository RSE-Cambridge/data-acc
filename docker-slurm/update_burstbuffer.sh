#!/bin/bash
set -eux

cd ../
make

cd docker-slurm
rm -rf ./bin
mkdir ./bin
cp ../bin/* ./bin

cp -r ../fs-ansible .

docker-compose down -v
docker-compose build
#docker-compose push
docker-compose up -d

docker exec slurmctld bash -c 'cd /data && echo "#!/bin/bash
#BB create_persistent name=mytestbuffer capacity=4000GB access=striped type=scratch" > create-persistent.sh'
docker exec slurmctld bash -c 'cd /data && echo "#!/bin/bash
#BB destroy_persistent name=mytestbuffer" > delete-persistent.sh'
docker exec slurmctld bash -c 'cd /data && echo "#!/bin/bash
#DW jobdw capacity=2TB access_mode=striped,private type=scratch
#DW persistentdw name=mytestbuffer
#DW swap 1GB
#DW stage_in source=/global/cscratch1/filename1 destination=\$DW_JOB_STRIPED/filename1 type=file
#DW stage_out source=\$DW_JOB_STRIPED/outdir destination=/global/scratch1/outdir type=directory
set
echo \$HOSTNAME
" > use-persistent.sh'


sleep 10
./register_cluster.sh

echo "Wait for startup to complete..."
sleep 10


echo "***Show current system state***"
docker exec slurmctld bash -c "cd /data && scontrol show burstbuffer"

sleep 3
echo "***Create persistent buffer***"
docker exec slurmctld bash -c "cd /data && cat create-persistent.sh"
docker exec slurmctld bash -c "cd /data && su slurm -c 'sbatch create-persistent.sh'"

sleep 5
docker exec slurmctld bash -c "cd /data && scontrol show burstbuffer"

echo "***Create per job buffer***"
echo 'srun --bb="capacity=1TB" bash -c "sleep 10 && echo \$HOSTNAME"'
docker exec slurmctld bash -c "cd /data && su slurm -c 'srun --bb=\"capacity=100GB\" bash -c \"sleep 5 && echo \$HOSTNAME\"'" &
sleep 5
docker exec slurmctld bash -c "cd /data && scontrol show burstbuffer"

sleep 5
echo "***Use persistent buffer***"
sleep 5
docker exec slurmctld bash -c "cd /data && cat use-persistent.sh"
docker exec slurmctld bash -c "cd /data && su slurm -c 'sbatch use-persistent.sh'"
sleep 5
docker exec slurmctld bash -c "cd /data && scontrol show burstbuffer"

echo "***Delete persistent buffer***"
docker exec slurmctld bash -c "cd /data && cat delete-persistent.sh"
docker exec slurmctld bash -c "cd /data && su slurm -c 'sbatch delete-persistent.sh'"

sleep 22
echo "***Show all buffers are cleaned up***"
docker exec slurmctld bash -c "cd /data && scontrol show burstbuffer"
echo "***Show all jobs completed***"
docker exec slurmctld bash -c "cd /data && squeue"

sleep 3
echo "***brick manager node 1 ***"
docker logs dockerslurm_fakebuffernode_1
echo "***brick manager node 2 ***"
docker logs dockerslurm_fakebuffernode_2
echo "***brick manager node 3 ***"
docker logs dockerslurm_fakebuffernode_3
sleep 3
echo "***Debugger: volumes ***"
docker logs dockerslurm_volumewatcher_1
echo "***Debugger: jobs ***"
docker logs dockerslurm_jobwatcher_1
echo "***Debugger: bricks ***"
docker logs dockerslurm_brickwatcher_1
sleep 3
echo "***Data still in etcd ***"
docker exec dockerslurm_brickwatcher_1 bash -c "etcdctl get --prefix /"

#sleep 15

# docker-compose logs -f &

# sleep 5

# echo For more details run "export ETCDCTL_API=3 watch --prefix buffers/"
# docker exec slurmctld bash -c "cd /data && su slurm -c 'srun --bb=\"capacity=1G\" bash -c \"set\"'"

# sleep 15

#
# Use this one to see logs as the job executes
#
# sleep 2
# docker exec slurmctld bash -c "cd /data && su slurm -c 'sbatch -n2 --bbf=buffer.txt --wrap=hostname'"
# docker logs slurmctld -f

# kill %1
