#!/bin/bash
set +eux

docker exec slurmctld bash -c 'cd /data && echo "#!/bin/bash
#BB create_persistent name=mytestbuffer capacity=2000GB access=striped type=scratch" > create-persistent.sh'
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
echo "***Debugger: volumes ***"
docker logs slurm-master_volumewatcher_1
echo "***Debugger: jobs ***"
docker logs slurm-master_jobwatcher_1
echo "***Debugger: bricks ***"
docker logs slurm-master_brickwatcher_1
sleep 3
echo "***Data still in etcd ***"
docker exec slurm-master_brickwatcher_1 bash -c "etcdctl get --prefix ''"

