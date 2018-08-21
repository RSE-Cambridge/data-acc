# configure beegfs for data-acc demo

To run this set of playbooks, please try these:

    ansible-playbook dac-b.yml -i inv-b.yml --tag format --tag mount
    ansible-playbook dac-b.yml -i inv-b.yml --tag mount
    ansible-playbook dac-b.yml -i inv-b.yml --tag create_mgs_mds
    ansible-playbook dac-b.yml -i inv-b.yml --tag create_mgs_mds
    ansible-playbook dac-b.yml -i inv-b.yml --tag create_storage
    ansible-playbook dac-b.yml -i inv-b.yml --tag create_storage
    ansible-playbook dac-b.yml -i inv-b.yml --tag stop
    ansible-playbook dac-b.yml -i inv-b.yml --tag stop
    ansible-playbook dac-b.yml -i inv-b.yml --tag create_mgs_mds
    ansible-playbook dac-b.yml -i inv-b.yml --tag create_storage
    ansible-playbook dac-b.yml -i inv-b.yml --tag stop
    ansible-playbook dac-b.yml -i inv-b.yml --tag unmount
    ansible-playbook dac-b.yml -i inv-b.yml --tag unmount
    ansible-playbook dac-b.yml -i inv-b.yml --tag format

## Install notes

You may find this useful to run the above ansible-playbook command:

    virtualenv .venv
    . .venv/bin/activate
    pip install -U pip
    pip install -U ansible
