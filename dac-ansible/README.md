# Install data-acc Development Environment with Ansible

Install data-acc demo with Ansible. There is a script to create OpenStack VMs
but you can skip OpenStack steps and use your own Ansible inventory file if
preferred.

## Setup

Install Ansible and the OpenStack SDK, eg in a Python virtual environment:

    virtualenv .venv
    . .venv/bin/activate
    pip install -U pip
    pip install -U ansible openstacksdk

Pull in Ansible role dependencies:

    ansible-galaxy install -r requirements.yml

Source your OpenStack RC, eg:

    . openrc

Create OpenStack VMs:

    ./create-servers.py -k KEYPAIR_NAME -n NETWORK_NAME > hosts
    
Once the VMs are created, you can now use Ansible to deploy the dev environment:

    ansible-playbook master.yml -i hosts

Once the Ansible has finished, you can login and try a Slurm test:

    ssh centos@<ip-of-slurm-master>
    sudo -i
    scontrol show burstbuffer
    /usr/local/bin/data-acc/tools/slurm-test.sh
    squeue
    scontrol show burstbuffer

## Debugging Guide

When trying slurm-test.sh or similar, here are some debugging tips.

### dac-slurm-master

Slurm master makes calls to dacctl via the datawarp burst buffer
plugin. This only really talks to etcd.

For slurmctld you can see the logs here:

    ssh centos@<ip-of-slurm-master>
    journalctld -u slurmctld

You can see the dacctl logs here:

    ssh centos@<ip-of-slurm-master>
    less /var/log/dacctl.log

When you have a buffer that needs to be teared down after fixing
what may have blocked any previous attempts (such as a bad sudoers files)
you can try:

    ssh centos@<ip-of-slurm-master>
    /usr/local/bin/dacctl teardown --token <job-id>

Note the above tends to leave client mounts behind, which need to be cleared
manually via "umount -l <directory>" on slurm-cpu[1-2].

### dac[1-3]

The dacd processes are listening to etcd waiting for commands from
dacctl via etcd. If they have the 0th brick, then they run Ansible
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
