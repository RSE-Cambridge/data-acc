#!/bin/bash
set -eux

cd /data
echo "#!/bin/bash
#BB create_persistent name=mytestbuffer capacity=4000GB access=striped type=scratch" > create-persistent.sh

echo "#!/bin/bash
#BB destroy_persistent name=mytestbuffer" > delete-persistent.sh

echo "#!/bin/bash
#DW jobdw capacity=2TB access_mode=striped,private type=scratch
#DW persistentdw name=mytestbuffer
#DW swap 1GB
#DW stage_in source=/global/cscratch1/filename1 destination=\$DW_JOB_STRIPED/filename1 type=file
#DW stage_out source=\$DW_JOB_STRIPED/outdir destination=/global/scratch1/outdir type=directory
set
echo \$HOSTNAME
" > use-persistent.sh

# Ensure Slurm is setup with the cluster name
/usr/bin/sacctmgr --immediate add cluster name=linux || true

SLEEP_INTERVAL=10

echo "Wait for startup to complete..."
sleep $SLEEP_INTERVAL

scontrol show burstbuffer

echo "***Create persistent buffer***"
cat create-persistent.sh
su slurm -c 'sbatch create-persistent.sh'

sleep $SLEEP_INTERVAL
scontrol show burstbuffer
squeue

echo "***Create per job buffer***"
su slurm -c 'srun --bb="capacity=100GB" bash -c "sleep 5 && echo \$HOSTNAME"'

sleep $SLEEP_INTERVAL
scontrol show burstbuffer
squeue
sleep $SLEEP_INTERVAL

echo "***Use persistent buffer***"
cat use-persistent.sh
su slurm -c 'sbatch use-persistent.sh'

sleep $SLEEP_INTERVAL
scontrol show burstbuffer
squeue
sleep $SLEEP_INTERVAL

echo "***Delete persistent buffer***"
cat delete-persistent.sh
su slurm -c 'sbatch delete-persistent.sh'
sleep $SLEEP_INTERVAL

echo "***Show all buffers are cleaned up***"
scontrol show burstbuffer
squeue
sleep $SLEEP_INTERVAL
scontrol show burstbuffer
squeue
