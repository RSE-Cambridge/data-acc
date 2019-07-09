# Install data-acc with Ansible

Install data-acc demo with ansible. It creates a bunch of OpenStack VMs, then uses ansible to install a development data-accellerator enviroment.

To run this set of playbooks, please execute:

    ./create-servers.py > hosts
    ansible-playbook master.yml -i hosts

## Install notes

You may find this useful to run the above ansible-playbook command:

    virtualenv .venv
    . .venv/bin/activate
    pip install -U pip
    pip install -U ansible openstacksdk
    ansible-galaxy install -r requirements.yml
    . openrc
