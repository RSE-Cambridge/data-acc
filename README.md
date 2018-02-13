# burstbuffer

Some experiments building a burst buffer, initially targeting Slurm users.

## Docker images

You can build the images like this:

  docker image build -t openhpc-base:v0.1 ./docker/openhpc-base
  docker image build -t openhpc-slurm-master:v0.1 ./docker/openhpc-slurm-master
  docker image build -t openhpc-slurm-compute:v0.1 ./docker/openhpc-slurm-compute

  docker network create slurm
  docker run -dt --privileged --hostname linux0 --name slurm-master \
      --network slurm -v /sys/fs/cgroup:/sys/fs/cgroup \
      openhpc-slurm-master:v0.1
  docker run -dt --privileged --hostname compute0 --name slurm-compute0 \
      --network slurm -v /sys/fs/cgroup:/sys/fs/cgroup \
      --add-host linux0 openhpc-slurm-compute:v0.1

  docker exec -it slurm-master /bin/bash

  docker image build -t openhpc-slurm-build:v0.1 ./docker/openhpc-slurm-build
  docker run -dt --privileged --name slurm-build \
      --network slurm -v /sys/fs/cgroup:/sys/fs/cgroup \
      openhpc-slurm-build:v0.1
  docker exec -it slurm-build /bin/bash
