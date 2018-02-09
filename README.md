# burstbuffer

Some experiments building a burst buffer, initially targeting Slurm users.

## Docker images

You can build the images like this:

  docker image build -t openhpc-slurm:v0.1 ./docker/openhpc-slurm
  docker run -d --name slurm openhpc-slurm:v0.1
  docker exec -it slurm /bin/bash
