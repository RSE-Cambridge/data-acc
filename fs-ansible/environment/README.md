# configure lustre for data-acc demo

To run this set of playbooks, please try:

    ansible-playbook test-dac.yml -i test-inventory

## Install notes

You may find this useful to run the above ansible-playbook command:

    virtualenv .venv
    . .venv/bin/activate
    pip install -U pip
    pip install -U ansible
