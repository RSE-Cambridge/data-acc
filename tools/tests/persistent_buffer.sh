#!/bin/bash

set -e

export name=test$RANDOM
scratch=$HOME/dac-tests-$name
output=$scratch/slurm.out

source "$( dirname $0 )/config"
source "$( dirname $0 )/functions"

create_files() {
  mkdir -p $scratch
  touch $scratch/$file

  echo "#!/bin/bash
#BB create_persistent name=$name capacity=$size access=striped type=scratch
#SBATCH --output=$output --open-mode=append
" > $scratch/create-persistent.sh

  echo "#!/bin/bash
#DW persistentdw name=$name
#DW stage_in source=$scratch/$file destination=\$DW_PERSISTENT_STRIPED_$name/$file type=file
#DW stage_out source=\$DW_PERSISTENT_STRIPED_$name/$file destination=$scratch/$outfile type=file
#SBATCH --output=$output --open-mode=append --nodes=$nodes
test -n \"\$DW_PERSISTENT_STRIPED_$name\" || echo environment variable is not set
test -d \"\$DW_PERSISTENT_STRIPED_$name\" || echo environment varibale is not a directory
test -f \"\$DW_PERSISTENT_STRIPED_$name/$file\" || echo stage in file $DW_PERSISTENT_STRIPED_$name/$file does not exist
echo \$SLURM_JOBID > \"\$DW_PERSISTENT_STRIPED_$name/$outfile\" || echo failed to write to file in buffer
if [[ \$( stat -c %U \$DW_PERSISTENT_STRIPED_$name/ ) != $USER ]]; then
  echo \"UID of \$DW_PERSISTENT_STRIPED_$name is not $USER\"; fi
if [[ \$( stat -c %G \$DW_PERSISTENT_STRIPED_$name/ ) != root ]]; then
  echo \"GID of \$DW_PERSISTENT_STRIPED_$name is not root\"; fi
if [[ \$( stat -c %a \$DW_PERSISTENT_STRIPED_$name/ ) != 700 ]]; then
  echo \"Permissions of \$DW_PERSISTENT_STRIPED_$name is not 700\"; fi" > $scratch/use-persistent.sh

  echo "#!/bin/bash
#BB destroy_persistent name=$name
#SBATCH --output=$output --open-mode=append
" > $scratch/delete-persistent.sh
}

submit_jobs() {
  sbatch --job-name=$name --quiet $scratch/create-persistent.sh
  sbatch --job-name=$name --quiet $scratch/use-persistent.sh
}

cleanup() {
  scancel -n $name -u $USER
  sbatch --job-name=$name --quiet $scratch/delete-persistent.sh
  rm -rf "$scratch"
  wait_for_jobs
}

source "$( dirname $0 )/template"
