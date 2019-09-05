#!/bin/bash
set -eux

cd /data
echo "#!/bin/bash
#BB create_persistent name=mytestbuffer capacity=4000GB access=striped type=scratch" > create-persistent.sh

echo "#!/bin/bash
#BB destroy_persistent name=mytestbuffer" > delete-persistent.sh

echo "#!/bin/bash
#DW jobdw capacity=2TB access_mode=striped,private type=scratch
#DW stage_in source=/usr/local/bin/dacd destination=\$DW_JOB_STRIPED/filename1 type=file
#DW stage_out source=\$DW_JOB_STRIPED/outdir destination=/tmp/perjob type=directory

env
df -h

mkdir \$DW_JOB_STRIPED/outdir
df -h > \$DW_JOB_STRIPED/outdir/dfoutput
ls -al \$DW_JOB_STRIPED > \$DW_JOB_STRIPED/outdir/lsoutput
file \$DW_JOB_STRIPED/filename1 > \$DW_JOB_STRIPED/outdir/stageinfile

echo \$HOSTNAME
" > use-perjob.sh

echo "#!/bin/bash
#DW persistentdw name=mytestbuffer
#DW stage_in source=/usr/local/bin/dacd destination=\$DW_PERSISTENT_STRIPED_mytestbuffer/filename1 type=file
#DW stage_out source=\$DW_PERSISTENT_STRIPED_mytestbuffer/outdir destination=/tmp/persistent type=directory

env
df -h

mkdir -p \$DW_PERSISTENT_STRIPED_mytestbuffer/outdir
echo \$SLURM_JOBID >> \$DW_PERSISTENT_STRIPED_mytestbuffer/outdir/jobids
ls -al \$DW_PERSISTENT_STRIPED_mytestbuffer >> \$DW_PERSISTENT_STRIPED_mytestbuffer/outdir/lsoutput
file \$DW_PERSISTENT_STRIPED_mytestbuffer/filename1 >> \$DW_PERSISTENT_STRIPED_mytestbuffer/outdir/stageinfile

echo \$HOSTNAME
" > use-persistent.sh

echo "#!/bin/bash
#DW jobdw capacity=2TB access_mode=striped,private type=scratch
#DW persistentdw name=mytestbuffer
#DW stage_in source=/usr/local/bin/dacd destination=\$DW_PERSISTENT_STRIPED_mytestbuffer/filename1 type=file
#DW stage_out source=\$DW_PERSISTENT_STRIPED_mytestbuffer/outdir destination=/tmp/persistent type=directory

env
df -h

touch \$DW_JOB_STRIPED/\$( date +%F )
ls -al \$DW_JOB_STRIPED

mkdir -p \$DW_PERSISTENT_STRIPED_mytestbuffer/outdir
echo \$SLURM_JOBID >> \$DW_PERSISTENT_STRIPED_mytestbuffer/outdir/jobids
ls -al \$DW_PERSISTENT_STRIPED_mytestbuffer >> \$DW_PERSISTENT_STRIPED_mytestbuffer/outdir/lsoutput
file \$DW_PERSISTENT_STRIPED_mytestbuffer/filename1 >> \$DW_PERSISTENT_STRIPED_mytestbuffer/outdir/stageinfile

echo \$HOSTNAME
" > use-multiple.sh

# Ensure Slurm is setup with the cluster name
/usr/bin/sacctmgr --immediate add cluster name=linux || true

SLEEP_INTERVAL=10

echo "Wait for startup to complete..."
sleep $SLEEP_INTERVAL

scontrol show burstbuffer
scontrol show bbstat

echo "***Create persistent buffer***"
cat create-persistent.sh
su slurm -c 'sbatch create-persistent.sh'

sleep $SLEEP_INTERVAL
scontrol show burstbuffer
squeue

echo "***Create per job buffer***"
su slurm -c 'srun --bb="capacity=100GB" bash -c "sleep 5 && echo \$HOSTNAME"'

scontrol show burstbuffer
squeue

echo "***Use persistent buffer***"
id centos &>/dev/null || adduser centos
cat use-multiple.sh
su centos -c 'sbatch --array=1-10 use-multiple.sh'
su centos -c 'sbatch use-multiple.sh'
su centos -c 'sbatch use-multiple.sh'
su centos -c 'sbatch use-multiple.sh'
su centos -c 'sbatch use-multiple.sh'
cat use-persistent.sh
su centos -c 'sbatch use-persistent.sh'
cat use-perjob.sh
su centos -c 'sbatch use-perjob.sh'

squeue

sleep $SLEEP_INTERVAL
scontrol show burstbuffer
squeue
sleep $SLEEP_INTERVAL

#echo "***Delete persistent buffer***"
#cat delete-persistent.sh
#su slurm -c 'sbatch delete-persistent.sh'
#sleep $SLEEP_INTERVAL

#echo "***Show all buffers are cleaned up***"
scontrol show burstbuffer
squeue
sleep $SLEEP_INTERVAL
scontrol show burstbuffer
squeue
scontrol show bbstat
