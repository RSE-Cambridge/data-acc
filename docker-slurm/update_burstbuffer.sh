#!/bin/bash
set -e

git push
docker exec slurmctld bash -c "cd /usr/local/src/burstbuffer && . .venv/bin/activate && git pull && pip install -Ue . && fakewarp help"
docker-compose restart slurmctld
