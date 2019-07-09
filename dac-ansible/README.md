# Install data-acc Development Environment with Ansible

Install data-acc demo with ansible. It creates a bunch of OpenStack VMs, then uses ansible to install a development data-acc enviroment.

To run this set of playbooks, please execute:

    . openrc
    ./create-servers.py > hosts
    ansible-playbook master.yml -i hosts

Note the above pulls the docker image johngarbutt/data-acc which can be
pushed by doing something like this:

    cd ../docker-slurm
    ./build.sh
    docker-compose push

## Install notes

You may find this useful to run the above ansible-playbook command:

    virtualenv .venv
    . .venv/bin/activate
    pip install -U pip
    pip install -U ansible openstacksdk
    ansible-galaxy install -r requirements.yml

## Debugging Guide

Once the ansible has finished, you can login and try a slurm test:

    ssh centos@<ip-of-slurm-master>
    docker exec -it slurmctld bash
    scontrol show burstbuffer
    cd /usr/local/bin/data-acc/tools/
    . slurm-test.sh

### dac-slurm-master

Slurm master makes calls to dacctl via the datawarp burst buffer
plugin. This only really talks to etcd.

For slurmctld you can see the logs here:

    ssh centos@<ip-of-slurm-master>
    docker logs slurmctld

You can see the dacctl logs here:

    ssh centos@<ip-of-slurm-master>
    docker exec -it slurmctld bash
    less /var/log/dacctl.log

When you have a buffer that needs to be teared down after fixing
what may have blocked any previous attempts (such as a bad sudoers files)
you can try:

    ssh centos@<ip-of-slurm-master>
    docker exec -it slurmctld bash
    /usr/local/bin/dacctl teardown --token <job-id>

### dac[1-3]

The dacd processes are listening to etcd waiting for commands from
dacctl via etcd. If they have the 0th brick, then they run ansible
over all the dacd nodes to create the filesystem, then they run ssh
to each of the compute nodes to mount the filesystem.

On the dacd nodes, you can find out lots from journalctld:

    ssh centos@<ip-of-dacd-node>
    journalctl -u dacd

You can also inspect the current state of data-acc by looking in etcd:

    ssh centos@<ip-of-dacd-node>
    sudo su dac /usr/local/bin/data-acc-v0.6/tools/etcd-ls.sh

You can check ssh access from dac by doing:

    ssh centos@<ip-of-dacd-node>
    ssh dac@dac1 date
    ssh dac@dac2 date
    ssh dac@dac3 date
    ssh dac@slurm-cpu1 date
    ssh dac@slurm-cpu2 date

### slurm-cpu[1-2]

Mostly watching for ssh from dacd that mounts lustre.

"mount" can give the current state of things, also with looking at
dmesg for lustre message.
