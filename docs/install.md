# Manual Installation of Data Accelerator

While there is an work in progress
[ansible based development setup](https://github.com/RSE-Cambridge/data-acc/blob/master/dac-ansible)
this document looks at the manual setup of the data-acc
components.

## Requirements

* Slurm 18.08.x or newer. 19.05.x is currently being tested
* etcd 3.3.x or newer
* Shared home directory between DAC, compute and login nodes

## DACD and DACCTL

Both software components are built statically.
There is a tagged release with a binary built using the CI system that is
ready to download from github:
[https://github.com/RSE-Cambridge/data-acc/releases](https://github.com/RSE-Cambridge/data-acc/releases)


The programs are statically compiled and can be copied around as per required.

The files are currently installed to `/usr/local/bin` on each DAC node.
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

On the Slurm master node, the `dacctl` binary needs to be accessible.
Below you can see the Slurm configuration options GetSysState and GetSysStatus,
both of which need to be modified to point to the location of the dacctl binary.

## etcd and TLS config

You need to install an etcd cluster.
It can be installed as required via EPEL or from
[the repository](https://www.github.com/coreos/etcd)

To secure the communication with etcd, TLS certificates should be used.

For more details please see the etcd docs, such as this page:
https://etcd.io/docs/v3.3.12/op-guide/maintenance/

## Example configuration

On each DAC node this environment file is used in the dacd unit file
to inform the dacd code the details of the environment, and in particular
have the correct certs needed to communicate with etcd:

**dacd**

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
pip install -U pip
pip install -U ansible
```

Note failed ansible runs are current left in /tmp.
This aids debugging, but left unchecked will consume all disk space.

The Slurm master unit file is updated to have this environment file
such that the dacctl command line can communicate with etcd:

**slurmctld**

```
ETCDCTL_API=3
ETCDCTL_ENDPOINTS=https://10.43.21.71:2379
ETCDCTL_CERT_FILE=/dac/dacd/cert/slurm-master.data.cluster.pem
ETCDCTL_KEY_FILE=/dac/dacd/cert/slurm-master.data.cluster-key.pem
ETCDCTL_CA_FILE=/dac/dacd/cert/ca.pem
```

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
