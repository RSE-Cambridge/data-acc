# configure beegfs for data-acc demo

To run this set of playbooks, please try these:

    ansible-playbook dac-b.yml -i inv-b.yml --tag format

## Install notes

You may find this useful to run the above ansible-playbook command:

    virtualenv .venv
    . .venv/bin/activate
    pip install -U pip
    pip install -U ansible
