# Configure fileystems for data-acc

This provides the ansible to configure Lustre file-systems for the
data accelerator. The BeeGFS scripts are currently not maintained.

The current entry points are the following playbooks:

* create.yml - create empty lustre filesystem
* delete.yml - teardown buffer, deleting partitions, disks not wiped
* restore.yml - re-mount filesystem (after dac host reboot)

The expected inventory format is best seen in the dac unit tests:

* https://github.com/RSE-Cambridge/data-acc/blob/master/internal/pkg/filesystem_impl/ansible_test.go

Note that dacctl is able to generate the ansible being used for any
current buffer via the `generate_ansible` command.

## Install notes

You may find this useful to run the above ansible-playbook command:

    virtualenv .venv
    . .venv/bin/activate
    pip install -U pip
    pip install -U ansible
