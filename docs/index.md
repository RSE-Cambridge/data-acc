# Data Accelerator

[![CircleCI](https://circleci.com/gh/RSE-Cambridge/data-acc.svg?style=svg&circle-token=4042ee71fb486efc320ce64b7b568afd4f9e0b38)](https://circleci.com/gh/RSE-Cambridge/data-acc)

The Data Accelerator (DAC) orchestrates the creation of burst buffers using
commodity hardware and existing parallel file systems.
Current focus is on creating NVMe backed Lustre file systems via
Slurm's Burst Buffer support.

https://rse-cambridge.github.io/data-acc/

The initial development focus has been on Cambridge University's Data
Accelerator. More information on the DAC can be found here:

* In June 2019 it reached #1 in the io500: https://www.vi4io.org/io500/start
* https://www.hpc.cam.ac.uk/research/data-acc
* The Cumulus supercomputer systems: https://www.top500.org/system/179577

Currently this makes use of 24 Dell EMC R740xd nodes.
Each contains two Intel OmniPath network adapters and 12 Intel P4600 SSDs.
Plans are in progress to support Mellanox HDR as well as other parallel file systems.

A whitepaper discussing the co-design between Cambridge, Intel, DellEMC and
StackHPC is available here:
https://www.dellemc.com/resources/en-us/asset/white-papers/products/ready-solutions/dell-data-accelerator-cambridge.pdf
This provides results and configuration information.

## CSD3 Data Accelerator Deployment

The initial target of this work is [CSD3](https://www.csd3.cam.ac.uk/) at the
University of Cambridge, and in particular
[Peta4](https://www.top500.org/system/179305).
Eventually, it is hoped to also connect it to
[Wilkes-2](https://www.top500.org/system/179044).

## Slurm Integration

A key part of the work is allowing users to request the burst buffer
resources using
[Slurm's burst buffer plugin](https://slurm.schedmd.com/burst_buffer.html).
Only the Cray DataWarp plugin is currently maintained, so integration is
focused on how to expose the data accelerator via the Cray DataWarp plugin.

There are two types of burst buffer:

* per job burst buffer
* persistent burst buffer

The user requests an amount of storage, the request is rounded up based on
the burst buffer pool granularity. Currently this is all done on a per
storage device granularity.

Initially we are targeting these access modes for per job burst buffers
(persistent burst buffers are always used as shared scratch):

1. Shared Scratch (all compute nodes access the same namespace,
   files are striped across all burst buffer nodes using PFS)
2. Private Scratch (separate namespace for each compute node)
3. Swap (additional swap file added on each assigned compute node)

In addition to orchestrating the creation of the buffer, there is
also an option to stage in and stage out files to and from an
existing slower storage tier. For example you might copy from
spinning disk based Lustre to an NVMe backed burst buffer.

The data accelerator project does not change how users interact with Slurm
requesting a burst buffer. NERSC have published a useful guide
on [how to use a burst buffer via slurm](http://www.nersc.gov/users/computational-systems/cori/burst-buffer/example-batch-scripts)

## Operator

There are two key binaries:

* **dacd** - runs on storage nodes, manages each host, watches keys in etcd
* **dacctl** - cli tool used by Slurm burst buffer plugin

To show what is added to a typical Slurm deployment when
you add the data accelerator [dac deployment diagram](https://drive.google.com/a/stackhpc.com/file/d/1UUrjlMtoyWETQuwdK1Pg0gyDe85GliGR/view?usp=sharing)

The overall data model of the system stored in etcd is covered in the following
[overall data model diagram](https://drive.google.com/a/stackhpc.com/file/d/1I3ot5pAc2-lID1w4JxFtD4bVPmeXuQ9Z/view?usp=sharing):

More details coming soon.

## Developer

For more details, such as how to build the golang code and run the unit tests,
please see: https://github.com/RSE-Cambridge/data-acc

### Slurm Integration Testing

Currently we are targeting testing at Slurm 19.05.x

The packing is based on:
https://github.com/giovtorres/slurm-docker-cluster

Currently uses the Fake PFS Provider, rather than using the fs-ansible repo
and its BeeGFS or Lustre support. (TODO, --dry-run or similar?)

For install with Slurm integration [please read the docker-slurm
README](https://github.com/RSE-Cambridge/data-acc/blob/master/docker-slurm/README.md)

#### Fake Demo

When using the [./update_burstbuffer.sh](https://github.com/RSE-Cambridge/data-acc/blob/master/docker-slurm/update_burstbuffer.sh)
script you get the following demo of the burst buffer:

```Console
***Show current system state***
Name=cray DefaultPool=dedicated_nvme Granularity=1TB TotalSpace=20TB FreeSpace=10TB UsedSpace=0
  Flags=EnablePersistent
  StageInTimeout=30 StageOutTimeout=30 ValidateTimeout=5 OtherTimeout=300
  AllowUsers=root,slurm
  GetSysState=/opt/cray/dw_wlm/default/bin/dw_wlm_cli
***Create persistent buffer***
#!/bin/bash
#BB create_persistent name=mytestbuffer capacity=32GB access=striped type=scratch
Submitted batch job 2
Name=cray DefaultPool=dedicated_nvme Granularity=1TB TotalSpace=20TB FreeSpace=9TB UsedSpace=1TB
  Flags=EnablePersistent
  StageInTimeout=30 StageOutTimeout=30 ValidateTimeout=5 OtherTimeout=300
  AllowUsers=root,slurm
  GetSysState=/opt/cray/dw_wlm/default/bin/dw_wlm_cli
  Allocated Buffers:
    Name=mytestbuffer CreateTime=2018-02-22T13:41:16 Pool=dedicated_nvme Size=1TB State=allocated UserID=slurm(995)
  Per User Buffer Use:
    UserID=slurm(995) Used=1TB
***Create per job buffer***
srun --bb="capacity=3TB" bash -c "sleep 10 && echo \$HOSTNAME"
srun: job 3 queued and waiting for resources
Name=cray DefaultPool=dedicated_nvme Granularity=1TB TotalSpace=20TB FreeSpace=6TB UsedSpace=4TB
  Flags=EnablePersistent
  StageInTimeout=30 StageOutTimeout=30 ValidateTimeout=5 OtherTimeout=300
  AllowUsers=root,slurm
  GetSysState=/opt/cray/dw_wlm/default/bin/dw_wlm_cli
  Allocated Buffers:
    JobID=3 CreateTime=2018-02-22T13:41:21 Pool=dedicated_nvme Size=3TB State=allocated UserID=slurm(995)
    Name=mytestbuffer CreateTime=2018-02-22T13:41:16 Pool=dedicated_nvme Size=1TB State=allocated UserID=slurm(995)
  Per User Buffer Use:
    UserID=slurm(995) Used=4TB
***Delete persistent buffer***
#!/bin/bash
#BB destroy_persistent name=mytestbuffer
Submitted batch job 4
srun: job 3 has been allocated resources
slurmctld
***Show all is cleaned up***
Name=cray DefaultPool=dedicated_nvme Granularity=1TB TotalSpace=20TB FreeSpace=13TB UsedSpace=0
  Flags=EnablePersistent
  StageInTimeout=30 StageOutTimeout=30 ValidateTimeout=5 OtherTimeout=300
  AllowUsers=root,slurm
  GetSysState=/opt/cray/dw_wlm/default/bin/dw_wlm_cli
```
