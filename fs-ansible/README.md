# configure fileystems for data-acc demo

For lustre we have:

    ansible-playbook test-dac-lustre.yml -i test-inventory --tag format
    ansible-playbook test-dac-lustre.yml -i test-inventory --tag mount,create_mdt,create_mgs,create_osts,client_mount
    ansible-playbook test-dac-lustre.yml -i test-inventory --tag mount,create_mdt,create_mgs,create_osts,client_mount
    ansible-playbook test-dac-beegfs.yml -i test-inventory --tag stop_all,unmount,client_unmount
    ansible-playbook test-dac-beegfs.yml -i test-inventory --tag stop_all,unmount,client_unmount
    ansible-playbook test-dac-lustre.yml -i test-inventory --tag format

    ansible-playbook test-dac-lustre.yml -i test-inventory --tag stop_mgs
    ansible-playbook test-dac-lustre.yml -i test-inventory --tag reformat_mgs


For beegfs we have:

    ansible-playbook test-dac-beegfs.yml -i test-inventory --tag format
    ansible-playbook test-dac-beegfs.yml -i test-inventory --tag mount --tag create_mdt --tag create_mgs --tag create_osts --tag client_mount
    ansible-playbook test-dac-beegfs.yml -i test-inventory --tag mount --tag create_mdt --tag create_mgs --tag create_osts --tag client_mount
    ansible-playbook test-dac-beegfs.yml -i test-inventory --tag stop_all --tag unmount --tag client_unmount
    ansible-playbook test-dac-beegfs.yml -i test-inventory --tag stop_all --tag unmount --tag client_unmount
    ansible-playbook test-dac-beegfs.yml -i test-inventory --tag format

## Install notes

You may find this useful to run the above ansible-playbook command:

    virtualenv .venv
    . .venv/bin/activate
    pip install -U pip
    pip install -U ansible
