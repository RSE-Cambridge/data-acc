# configure lustre for data-acc demo

To run this set of playbooks, please try these:

    ansible-playbook test-dac.yml -i test-inventory --tag format_mgs --tag format_mdts --tag format_osts
    ansible-playbook test-dac.yml -i test-inventory --tag start_mgs --tag start_mdts --tag start_osts --tag mount_fs
    ansible-playbook test-dac.yml -i test-inventory --tag start_mgs --tag start_mdts --tag start_osts --tag mount_fs

    ansible-playbook test-dac.yml -i test-inventory --tag umount_fs --tag stop_osts --tag stop_mdts
    ansible-playbook test-dac.yml -i test-inventory --tag umount_fs --tag stop_osts --tag stop_mdts
    ansible-playbook test-dac.yml -i test-inventory --tag format_mdts --tag format_osts

    ansible-playbook test-dac.yml -i test-inventory --tag stop_mgs
    ansible-playbook test-dac.yml -i test-inventory --tag reformat_mgs

## Install notes

You may find this useful to run the above ansible-playbook command:

    virtualenv .venv
    . .venv/bin/activate
    pip install -U pip
    pip install -U ansible
