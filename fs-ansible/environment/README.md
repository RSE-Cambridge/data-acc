# configure lustre for data-acc demo

To run this set of playbooks, please try these:

    ansible-playbook test-dac.yml -i test-inventory --tag format_mgs --tag reformat_mdts --tag reformat_osts
    ansible-playbook test-dac2.yml -i test-inventory2 --tag format_mgs --tag reformat_mdts --tag reformat_osts

    ansible-playbook test-dac.yml -i test-inventory --tag start_mgs --tag start_mdts --tag start_osts --tag mount_fs
    ansible-playbook test-dac2.yml -i test-inventory2 --tag start_mgs --tag start_mdts --tag start_osts --tag mount_fs
    ansible-playbook test-dac.yml -i test-inventory --tag start_mgs --tag start_mdts --tag start_osts --tag mount_fs

    ansible-playbook test-dac.yml -i test-inventory --tag umount_fs --tag stop_osts --tag stop_mdts
    ansible-playbook test-dac2.yml -i test-inventory2 --tag umount_fs --tag stop_osts --tag stop_mdts
    ansible-playbook test-dac2.yml -i test-inventory2 --tag umount_fs --tag stop_osts --tag stop_mdts

    ansible-playbook test-dac.yml -i test-inventory --tag reformat_mdts --tag reformat_osts
    ansible-playbook test-dac2.yml -i test-inventory2 --tag reformat_mdts --tag reformat_osts

    ansible-playbook test-dac.yml -i test-inventory --tag stop_mgs
    ansible-playbook test-dac.yml -i test-inventory --tag reformat_mgs


For beegfs we have:

    ansible-playbook test-dac-beegfs.yml -i test-inventory --tag format
    ansible-playbook test-dac-beegfs-2.yml -i test-inventory2 --tag format

## Install notes

You may find this useful to run the above ansible-playbook command:

    virtualenv .venv
    . .venv/bin/activate
    pip install -U pip
    pip install -U ansible
