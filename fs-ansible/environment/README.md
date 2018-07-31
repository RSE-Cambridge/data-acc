# configure lustre for data-acc demo

To run this set of playbooks, please try these:

    ansible-playbook test-dac.yml -i test-inventory --tag format_mdtmgs --tag format_osts
    ansible-playbook test-dac.yml -i test-inventory --tag start_osts --tag start_mgsdt --tag mount_fs

    ansible-playbook test-dac.yml -i test-inventory --tag stop_osts --tag stop_mgsdt --tag umount_fs
    ansible-playbook test-dac.yml -i test-inventory --tag format_mdtmgs --tag format_osts

## Install notes

You may find this useful to run the above ansible-playbook command:

    virtualenv .venv
    . .venv/bin/activate
    pip install -U pip
    pip install -U ansible
