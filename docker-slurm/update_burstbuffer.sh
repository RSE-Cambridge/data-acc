#!/bin/bash
set +eux

cd ../
make
cd docker-slurm
mkdir ./bin
cp ../bin/amd64/* ./bin
docker build -t slurm-docker-cluster:17.02.9 .

#docker exec slurmctld bash -c "cd /usr/local/src/burstbuffer && . .venv/bin/activate && git remote update && git checkout etcd && git pull && pip install -Ue . && fakewarp help"

docker-compose up -d

sleep 8
./register_cluster.sh

#sleep 5
#docker exec bufferwatcher bash -c "data-acc-host"

docker exec slurmctld bash -c 'cd /data && echo "#!/bin/bash
#BB create_persistent name=mytestbuffer capacity=32GB access=striped type=scratch" > create-persistent.sh'
docker exec slurmctld bash -c 'cd /data && echo "#!/bin/bash
#BB destroy_persistent name=mytestbuffer" > delete-persistent.sh'

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
docker exec slurmctld bash -c "cd /data && su slurm -c 'srun --bb=\"capacity=1TB\" bash -c \"sleep 5 && echo \$HOSTNAME\"'" &
sleep 5
docker exec slurmctld bash -c "cd /data && scontrol show burstbuffer"

sleep 5
echo "***Delete persistent buffer***"
docker exec slurmctld bash -c "cd /data && cat delete-persistent.sh"
docker exec slurmctld bash -c "cd /data && su slurm -c 'sbatch delete-persistent.sh'"

sleep 22
echo "***Show all is cleaned up***"
docker exec slurmctld bash -c "cd /data && scontrol show burstbuffer"
docker exec slurmctld bash -c "cd /data && squeue"

sleep 3
docker logs dockerslurm_fakebuffernode_1
sleep 3
docker logs bufferwatcher
sleep 3
docker exec bufferwatcher bash -c "etcdctl get --prefix /"

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
