# Manual Instalation of Data Accelerator Orchestrator

This follows a manual setup of the softare and some of the setings used on Cumulus to support the
testing. Upgrades and improvements are ongoing. An ansible enabled install is provided in
dac-ansible. This guide should serve as the a explanation of what is required and this can be
automated as per the requirements of each centers procedures.


## DACD and DACCTL

Both software components are built statically. There is a tagged release with a built binry ready to
download from github. The programs are statically compiled and can be copyed as per required. Thease
files are currently installed to /usr/local/bin on each DAC node and dacctl installed on the slurm
master.


The following systemd unit file is used to control the program

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
## ETCD
This can be installed as required via EPEL or from [the repository](https://www.github.com/coreos/etcd)



## DACD Configuration and TLS CERTS
In order for the software to communicate with ETCD each DAC node will need a TLS CERT. If sutch infrastructure dose not exist dac-ansible can be used to create the required certs. This makes use of the automated TLS generation scrips provided by CloudFlare.

The following config files for dacd on each DAC node and slurmctld on the slurm master need configuring. Thease configuration files are stored on a 

**dacd**

```
ETCDCTL_API=3
ETCDCTL_ENDPOINTS=https://10.43.21.71:2379 ETCD_CLIENT_DEBUG=1
ETCDCTL_CERT_FILE=/dac/dacd/cert/dac-e-24.data.cluster.pem
ETCDCTL_KEY_FILE=/dac/dacd/cert/dac-e-24.data.cluster-key.pem
ETCDCTL_CA_FILE=/dac/dacd/cert/ca.pem
```

**slurmctld**

```
ETCDCTL_API=3
ETCDCTL_ENDPOINTS=https://10.43.21.71:2379
ETCDCTL_CERT_FILE=/dac/dacd/cert/slurm-master.data.cluster.pem
ETCDCTL_KEY_FILE=/dac/dacd/cert/slurm-master.data.cluster-key.pem
ETCDCTL_CA_FILE=/dac/dacd/cert/ca.pem
```

## Slurm Configuration
Add the following to or crate the files 
**slurm.conf**

```
BurstBufferType=burst_buffer/cray
```

**burst_buffer.conf**

```
AllowUsers=root,slurm
Flags=EnablePersistent,SetExecHost,PrivateData
 
StageInTimeout=3600
StageOutTimeout=3600
OtherTimeout=300
 
DefaultPool=default
 
GetSysState=/usr/local/bin/dacctl
```

## SSH and Service Accounts
As ansible is used to mount the filesystem on each compute node, SSH keys, knowen host and authorized keys should be consistent across the infrastructure. We are currently looking at alternetives sutch as Kerberos or replace this step with a futher slurm plugin.

**dacd** is given a service account at cambridge to perform the on compute node acctions. The deamon on the DAC nodes is also run as this used. On the compute nodes we defined DACD with sudoers accsess to the following.

```
dacd ALL=(ALL) NOPASSWD: /usr/bin/mkdir -p /dac/*, /usr/bin/chmod 770 /dac/*, /usr/bin/chmod 0600 /dac/*, /usr/bin/chown * /dac/*, /usr/bin/mount -t lustre * /dac/*, /usr/bin/umount -l /dac/*, /usr/bin/umount /dac/*, /usr/sbin/losetup /dev/loop* /dac/*, /usr/sbin/losetup -d /dev/loop*, /usr/sbin/mkswap /dev/loop*, /usr/sbin/swapon /dev/loop*, /usr/sbin/swapoff /dev/loop*, /usr/bin/ln -s /dac/* /dac/*, /usr/bin/dd if=/dev/zero of=/dac/*, /usr/bin/rm -rf /dac/*, /bin/grep /dac/* /etc/mtab
```
This restricts the dac to use only that which is requierd and minipulate files under */dac/*