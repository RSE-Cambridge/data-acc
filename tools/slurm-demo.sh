#!/bin/bash
# Create and run jobs as the current user to test per-job and persistent buffers
# including copy in and out.

set -eu

mkdir -p ~/slurm-jobs
cd ~/slurm-jobs
touch perjob.txt persistent.txt

echo "#!/bin/bash
#BB create_persistent name=${USER}buffer capacity=1TB access=striped type=scratch" > create-persistent.sh

echo "#!/bin/bash
#BB destroy_persistent name=${USER}buffer" > delete-persistent.sh

echo "#!/bin/bash
#DW jobdw capacity=1TB access_mode=striped type=scratch
#DW stage_in source=$HOME/slurm-jobs/perjob.txt destination=\$DW_JOB_STRIPED/perjob.txt type=file
#DW stage_out source=\$DW_JOB_STRIPED/perjob.txt destination=$HOME/slurm-jobs/perjob.txt type=file

date >> \$DW_JOB_STRIPED/perjob.txt
echo \$HOSTNAME >> \$DW_JOB_STRIPED/perjob.txt
echo \$SLURM_JOBID >> \$DW_JOB_STRIPED/perjob.txt" > use-perjob.sh

echo "#!/bin/bash
#DW persistentdw name=${USER}buffer
#DW stage_in source=$HOME/slurm-jobs/persistent.txt destination=\$DW_PERSISTENT_STRIPED_${USER}buffer/persistent.txt type=file
#DW stage_out source=\$DW_PERSISTENT_STRIPED_${USER}buffer/persistent.txt destination=$HOME/slurm-jobs/persistent.txt type=file

date >> \$DW_PERSISTENT_STRIPED_${USER}buffer/persistent.txt
echo \$HOSTNAME >> \$DW_PERSISTENT_STRIPED_${USER}buffer/persistent.txt
echo \$SLURM_JOBID >> \$DW_PERSISTENT_STRIPED_${USER}buffer/persistent.txt" > use-persistent.sh

wait_for_jobs() {
  echo "***Waiting for jobs to complete***"
  until [ -z "$(squeue -h -u $USER)" ]; do
    QUEUE=$( squeue -u $USER )
    LINES=$( echo "$QUEUE" | wc -l )
    echo "$QUEUE" | cut -c 1-$( tput cols )
    sleep 10
    while [ $LINES -gt 0 ]; do
      tput cuu1
      ((LINES--))
    done
    tput ed
  done
}

cleanup() {
  scancel -u $USER
  sbatch delete-persistent.sh
  wait_for_jobs
}

trap cleanup EXIT

echo "***Current buffers visible by $USER***"
scontrol show burstbuffer

echo "***Run jobs***"
sbatch create-persistent.sh
sbatch use-perjob.sh
sbatch use-persistent.sh
wait_for_jobs

echo "***Run a second job to use the persistent buffer***"
sbatch use-persistent.sh
wait_for_jobs

echo "***Delete persistent buffer***"
cleanup
trap - EXIT

echo "***Show if all buffers are cleaned up***"
scontrol show burstbuffer

echo "***Read job data***"
echo "Per job:"
echo "========"
cat perjob.txt
echo ""
echo "Persistent:"
echo "==========="
cat persistent.txt
