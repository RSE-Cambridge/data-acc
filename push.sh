scp ./bin/*.tgz jmfg2@login-cpu.hpc.cam.ac.uk:~/

cd docker-slurm
#./build.sh
docker-compose push
