# Manual Installation of Data Accelerator

While there is an work in progress
[ansible based development setup](https://github.com/RSE-Cambridge/data-acc/blob/master/dac-ansible)
this document looks at the manual setup of the data-acc
components.

End users request burst buffers from Slurm using the interface defined here:
https://slurm.schedmd.com/burst_buffer.html

Every node providing storage runs a dacd process. The slurm master makes use of
dacctl command line tool, that works with the Slurm burst buffer plugin.
They communicate via writing and reading key value pairs in etcd.
Etcd is a key value store, for more details see: https://etcd.io/

## Requirements

* Slurm 19.05.x (18.08 possibly works)
* Lustre 2.12.x or newer with LNET configured
* etcd 3.3.x or newer
* Python 2 (Python 3 is not supported)
* Shared home directory between DAC, compute and login nodes

## DACD and DACCTL

Both software components are built statically, as is the norm for golang.
There is a tagged release with a binary built using the CI system that is
ready to download from github:
[https://github.com/RSE-Cambridge/data-acc/releases](https://github.com/RSE-Cambridge/data-acc/releases)

The dacd and dacctl binaries are typically copied from the above release tarball into:
`/usr/local/bin` on both the slurm master and the dacd server

All their configuration comes from environement variables.
Typically these are set using an EnvironmentFile inside the appropriate systemd unit file.
For example, for dacctl, it is run by slurmctld, so you need to set the environment variables
for the slurmctld process, such that they are passed to dacctl.

Example configration will follow in this guide, but for more details on what
configuration options are supported, please see the code:
https://github.com/RSE-Cambridge/data-acc/tree/v2.4/internal/pkg/config

## etcd and TLS config

You need to install an etcd cluster.
It can be installed as required via EPEL or from
[the repository](https://www.github.com/coreos/etcd)
But really we 

To secure the communication with etcd, TLS certificates should be used.

For more details please see the etcd docs, such as this page:
https://etcd.io/docs/v3.3.12/op-guide/maintenance/

## Example configuration

On each DAC node this environment file is used in the dacd unit file
to inform the dacd code the details of the environment, and in particular
have the correct certs needed to communicate with etcd:

**dacd**

The following systemd unit file is used to control the dacd program:

```
[Unit]
Description=dacd
ConditionPathExists=/usr/local/bin/dacd
After=network.target

[Service]
Type=simple
User=dacd
Group=dacd
Restart=on-failure
RestartSec=10

WorkingDirectory=/var/lib/data-acc
EnvironmentFile=/etc/dacd/dacd.conf
ExecStart=/usr/local/bin/dacd

# make sure log directory exists and owned by syslog
PermissionsStartOnly=true
StandardOutput=syslog
StandardError=syslog
SyslogIdentifier=dacd

[Install]
WantedBy=multi-user.target
```

The configuration in `/etc/dacd/dacd.conf` is covered in more detail below.

```
ETCDCTL_API=3
ETCDCTL_ENDPOINTS=https://10.43.21.71:2379
ETCD_CLIENT_DEBUG=1
ETCDCTL_CERT_FILE=/dac/dacd/cert/dac-e-24.data.cluster.pem
ETCDCTL_KEY_FILE=/dac/dacd/cert/dac-e-24.data.cluster-key.pem
ETCDCTL_CA_FILE=/dac/dacd/cert/ca.pem
DAC_LNET_SUFFIX="-opa@o2ib1"
DEVICE_COUNT=12
DAC_DEVICE_CAPACITY_GB=1400
DAC_POOL_NAME=default
DAC_MGS_DEV=sdb
DAC_HOST_GROUP=dac-prod
DAC_SKIP_ANSIBLE=false
DAC_MDT_SIZE="20g"
DAC_ANSIBLE_DIR=/var/lib/data-acc/fs-ansible/
```

Note that `/var/lib/data-acc/fs-ansible/` should contain the fs-ansible
scripts from the release tarball.
In addition there should be a working virtual environment in
`/var/lib/data-acc/fs-ansible/.venv`.
You can can create this on each dacd node as follows:

```
virtualenv /var/lib/data-acc/fs-ansible/.venv
. /var/lib/data-acc/fs-ansible/.venv/bin/activate
python --version # Ensure this is python 2
pip install -U pip
pip install -U ansible
```

Note failed ansible runs are current left in /tmp.
This aids debugging, but left unchecked will consume all disk space.

**slurmctld**

On the Slurm master node, the `dacctl` binary needs to be accessible and
/var/log/dacctl.log needs to be writable by the slurm user.

dacctl needs configuration so it knows how to contact the dacctl binary.
The Slurm master unit file is updated to have this environment file
such that the dacctl command line can communicate with etcd:

```
ETCDCTL_API=3
ETCDCTL_ENDPOINTS=https://10.43.21.71:2379
ETCDCTL_CERT_FILE=/dac/dacd/cert/slurm-master.data.cluster.pem
ETCDCTL_KEY_FILE=/dac/dacd/cert/slurm-master.data.cluster-key.pem
ETCDCTL_CA_FILE=/dac/dacd/cert/ca.pem
```

Below you can see the Slurm configuration options GetSysState and GetSysStatus,
both of which need to be modified to point to the location of the dacctl binary.

## Slurm Configuration

Here are import parts of the Slurm configuration files
that need merging with your existing Slurm configuration:

**slurm.conf**

```
BurstBufferType=burst_buffer/cray
AccountingStorageTRES="bb/cray"
```

From Slurm 19.04 you instead put this:

```
BurstBufferType=burst_buffer/datawarp
AccountingStorageTRES="bb/datawarp"
```

**burst_buffer.conf**

```
AllowUsers=root,slurm
Flags=EnablePersistent,PrivateData
 
StageInTimeout=3600
StageOutTimeout=3600
OtherTimeout=300
 
DefaultPool=default
 
GetSysState=/usr/local/bin/dacctl
GetSysStatus=/usr/local/bin/dacctl
```

## SSH and Service Accounts

The `dacd` daemon makes use of Ansible to create the filesystem. It also
uses ssh to then mount the file system on the compute nodes.
As such, the user that is running `dacd` needs to be able to ssh to all
the servers running `dacd` and all the compute nodes, using the hostnames
recorded in etcd and Slurm respectively.

For running Ansible, password-less sudo is needed for the user across
all the nodes running `dacd`.
On the compute nodes, it seems possible to restrict sudo access to the
following:
```
dacd ALL=(ALL) NOPASSWD: /usr/bin/mkdir -p /mnt/dac/*, /usr/bin/chmod 700 /mnt/dac/*, /usr/bin/chmod 0600 /mnt/dac/*, /usr/bin/chown * /mnt/dac/*, /usr/bin/mount -t lustre * /mnt/dac/*, /usr/bin/umount /mnt/dac/*, /usr/sbin/losetup /dev/loop* /mnt/dac/*, /usr/sbin/losetup -d /dev/loop*, /usr/sbin/mkswap /dev/loop*, /usr/sbin/swapon /dev/loop*, /usr/sbin/swapoff /dev/loop*, /usr/bin/ln -s /mnt/dac/* /mnt/dac/*, /usr/bin/dd if=/dev/zero of=/dac/*, /usr/bin/rm -df /mnt/dac/*, /bin/grep /mnt/dac/* /etc/mtab
```
