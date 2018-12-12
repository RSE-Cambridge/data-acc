#make clean
#make
#make tar
scp `ls ./bin/*.tgz` jmfg2@login-cpu.hpc.cam.ac.uk:~/data-acc-v0.5.tgz

cd docker-slurm
#./build.sh
docker-compose push
