#!/bin/bash

set -e

export name=test$RANDOM
scratch=$HOME/dac-tests-$name
output=$scratch/slurm.out

source "$( dirname $0 )/config"
source "$( dirname $0 )/functions"

# Calculate the size of the buffers to use, we'll create two each needing 2/3 of the space
export $( scontrol show bu | head -n1 )
size=$( expr ${TotalSpace//[!0-9]/} / 3 \* 2 )${TotalSpace//[0-9]/}

create_files() {
  mkdir -p $scratch
  touch $scratch/$file

  echo "#!/bin/bash
#DW jobdw capacity=$size access_mode=striped type=scratch
#DW stage_in source=$scratch/$file destination=\$DW_JOB_STRIPED/$file type=file
#DW stage_out source=\$DW_JOB_STRIPED/$file destination=$scratch/$outfile type=file
#SBATCH --output=$output --open-mode=append
test -n \"\$DW_JOB_STRIPED\" || echo 'environment variable \$DW_JOB_STRIPED is not set'
test -d \"\$DW_JOB_STRIPED\" || echo 'environment variable \$DW_JOB_STRIPED is not a directory'
touch \"\$DW_JOB_STRIPED/$file\" || echo 'failed to touch file in \$DW_JOB_STRIPED'" > $scratch/over-allocate.sh
}

submit_jobs() {
  for num in {1..2}; do
    sbatch --job-name=$name --quiet $scratch/over-allocate.sh
  done
}

cleanup() {
  scancel -n $name -u $USER
  rm -rf "$scratch"
  wait_for_jobs
}

source "$( dirname $0 )/template"
