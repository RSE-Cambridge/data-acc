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
#DW jobdw capacity=$size access_mode=private type=scratch
#SBATCH --output=$output --open-mode=append --nodes=$nodes
set -e
test -n \"\$DW_JOB_PRIVATE\" || echo 'environment variable \$DW_JOB_PRIVATE is not set'
test -d \"\$DW_JOB_PRIVATE\" || echo 'environment variable \$DW_JOB_PRIVATE is not a directory'
touch \"\$DW_JOB_PRIVATE/$file\" || echo 'failed to touch file in \$DW_JOB_PRIVATE'
stat_uid=\$( stat -c %U \$DW_JOB_PRIVATE/ )
stat_gid=\$( stat -c %G \$DW_JOB_PRIVATE/ )
stat_perm=\$( stat -c %a \$DW_JOB_PRIVATE/ )
if [[ \$stat_uid != $USER ]]; then
  echo \"UID of \$DW_JOB_PRIVATE is \$stat_uid and not $USER\"; fi
if [[ \$stat_gid != $USER ]]; then
  echo \"GID of \$DW_JOB_PRIVATE is \$stat_gid and not $USER\"; fi
if [[ \$stat_perm != 700 ]]; then
  echo \"Permissions of \$DW_JOB_PRIVATE are \$stat_perm and not 700\"; fi" > $scratch/use-private.sh


  echo "#!/bin/bash
#DW jobdw capacity=$size access_mode=striped type=scratch
#DW stage_in source=$scratch/$file destination=\$DW_JOB_STRIPED/$file type=file
#DW stage_out source=\$DW_JOB_STRIPED/$file destination=$scratch/$outfile type=file
#SBATCH --output=$output --open-mode=append --nodes=$nodes
set -e
test -n \"\$DW_JOB_STRIPED\" || echo 'environment variable \$DW_JOB_STRIPED is not set'
test -d \"\$DW_JOB_STRIPED\" || echo 'environment variable \$DW_JOB_STRIPED is not a directory'
test -f \"\$DW_JOB_STRIPED/$file\" || echo 'stage in file \$DW_JOB_STRIPED/$file does not exist'
touch \"\$DW_JOB_STRIPED/$file\" || echo 'failed to touch file in \$DW_JOB_STRIPED'
stat_uid=\$( stat -c %U \$DW_JOB_STRIPED/ )
stat_gid=\$( stat -c %G \$DW_JOB_STRIPED/ )
stat_perm=\$( stat -c %a \$DW_JOB_STRIPED/ )
if [[ \$stat_uid != $USER ]]; then
  echo \"UID of \$DW_JOB_STRIPED is \$stat_uid and not $USER\"; fi
if [[ \$stat_gid != $USER ]]; then
  echo \"GID of \$DW_JOB_STRIPED is \$stat_gid and not $USER\"; fi
if [[ \$stat_perm != 700 ]]; then
  echo \"Permissions of \$DW_JOB_STRIPED are \$stat_perm and not 700\"; fi" > $scratch/use-striped.sh

  echo "#!/bin/bash
#DW jobdw capacity=$size access_mode=private,striped type=scratch
#DW stage_in source=$scratch/$file destination=\$DW_JOB_STRIPED/$file type=file
#DW stage_out source=\$DW_JOB_STRIPED/$file destination=$scratch/$file type=file
#SBATCH --output=$output --open-mode=append --nodes=$nodes
set -e
test -n \"\$DW_JOB_PRIVATE\" || echo 'environment variable \$DW_JOB_PRIVATE is not set'
test -n \"\$DW_JOB_STRIPED\" || echo 'environment variable \$DW_JOB_STRIPED is not set'
test -d \"\$DW_JOB_PRIVATE\" || echo 'environment variable \$DW_JOB_PRIVATE is not a directory'
test -d \"\$DW_JOB_STRIPED\" || echo 'environment variable \$DW_JOB_STRIPED is not a directory'
test -f \"\$DW_JOB_STRIPED/$file\" || echo 'stage in file \$DW_JOB_STRIPED/$file does not exist'
touch \"\$DW_JOB_PRIVATE/$file\" || echo 'failed to touch file in \$DW_JOB_PRIVATE'
touch \"\$DW_JOB_STRIPED/$file\" || echo 'failed to touch file in \$DW_JOB_STRIPED'
stat_uid=\$( stat -c %U \$DW_JOB_STRIPED/ )
stat_gid=\$( stat -c %G \$DW_JOB_STRIPED/ )
stat_perm=\$( stat -c %a \$DW_JOB_STRIPED/ )
if [[ \$stat_uid != $USER ]]; then
  echo \"UID of \$DW_JOB_STRIPED is \$stat_uid and not $USER\"; fi
if [[ \$stat_gid != $USER ]]; then
  echo \"GID of \$DW_JOB_STRIPED is \$stat_gid and not $USER\"; fi
if [[ \$stat_perm != 700 ]]; then
  echo \"Permissions of \$DW_JOB_STRIPED are \$stat_perm and not 700\"; fi
stat_uid=\$( stat -c %U \$DW_JOB_PRIVATE/ )
stat_gid=\$( stat -c %G \$DW_JOB_PRIVATE/ )
stat_perm=\$( stat -c %a \$DW_JOB_PRIVATE/ )
if [[ \$stat_uid != $USER ]]; then
  echo \"UID of \$DW_JOB_PRIVATE is \$stat_uid and not $USER\"; fi
if [[ \$stat_gid != $USER ]]; then
  echo \"GID of \$DW_JOB_PRIVATE is \$stat_gid and not $USER\"; fi
if [[ \$stat_perm != 700 ]]; then
  echo \"Permissions of \$DW_JOB_PRIVATE are \$stat_perm and not 700\"; fi" > $scratch/use-both.sh
}

submit_jobs() {
  sbatch --job-name=$name --quiet $scratch/use-striped.sh
  sbatch --job-name=$name --quiet $scratch/use-private.sh
  sbatch --job-name=$name --quiet $scratch/use-both.sh
  sbatch --job-name=$name --quiet --array=1-10 $scratch/use-striped.sh
  sbatch --job-name=$name --quiet --array=1-10 $scratch/use-private.sh
  sbatch --job-name=$name --quiet --array=1-10 $scratch/use-both.sh
}

cleanup() {
  scancel -n $name -u $USER
  rm -rf "$scratch"
  wait_for_jobs
}

source "$( dirname $0 )/template"
