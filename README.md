# Data Accelerator

[![CircleCI](https://circleci.com/gh/RSE-Cambridge/data-acc.svg?style=svg&circle-token=4042ee71fb486efc320ce64b7b568afd4f9e0b38)](https://circleci.com/gh/RSE-Cambridge/data-acc)
[![Go Report Card](https://goreportcard.com/badge/github.com/RSE-Cambridge/data-acc)](https://goreportcard.com/report/github.com/RSE-Cambridge/data-acc)
[![codecov](https://codecov.io/gh/RSE-Cambridge/data-acc/branch/master/graph/badge.svg)](https://codecov.io/gh/RSE-Cambridge/data-acc)
[![Releases](https://img.shields.io/github/release/RSE-Cambridge/data-acc/all.svg?style=flat-square)](https://github.com/RSE-Cambridge/data-acc/releases)
[![LICENSE](https://img.shields.io/github/license/RSE-Cambridge/data-acc.svg?style=flat-square)](https://github.com/RSE-Cambridge/data-acc/blob/master/LICENSE)
<!-- 
[![Godoc](http://img.shields.io/badge/go-documentation-blue.svg?style=flat-square)](https://godoc.org/github.com/RSE-Cambridge/data-acc)
[![Build Status](https://travis-ci.org/RSE-Cambridge/data-acc.svg?branch=master)](https://travis-ci.org/RSE-Cambridge/data-acc)
-->

The Data Accelerator orchestrates the creation of burst buffers using
commodity hardware and existing parallel file systems.
Current focus is on creating NVMe backed Lustre file systems via
Slurm's Burst Buffer support.

https://rse-cambridge.github.io/data-acc/

The initial development focus has been on Cambridge University's Data Accelerator:

* In June 2019 it got #1 in io500: https://www.vi4io.org/io500/start
* https://www.hpc.cam.ac.uk/research/data-acc
* Attached to the Cumulus supercomputer: https://www.top500.org/system/179577

Currently this makes use of 24 Dell EMC R740xd nodes.
Each contains two Intel OmniPath network adapters and 12 Intel P4600 SSDs.

## Try me in docker-compose

We have a Docker Compose based Integration Test so you can try out how
we integrated with Slurm.
To see an end to end demo with Slurm 19.04
(but without running fs-ansible and not ssh-ing to compute nodes to mount)
please try:
```
cd docker-slurm
./demo.sh
```

To clean up after the demo, including removing all docker volumes:
```
docker-compose down --vol --rmi all
```

For more details please see the
[docker compose README](docker-slurm/README.md).

### Other Installation Guides

For an Ansible driven deployment into OpenStack VMs
(useful for testing out the ansble that creates Lustre filesystems on demand),
please take a look at:
[Development Environment Install Guide](dac-ansible/)

For a manual install there are some pointers in:
[Manual Install Guide](docs/install.md)

## How it works

There are two key binaries produced by the golang based code:

* [dacd](cmd/dacd): service that runs on the storage nodes to orchestrate filesystem creation
* [dacctl](cmd/dacctl): CLI tool used by Slurm Cray DataWarp burst buffer plugin to orchestration burst buffer creation

All the dacd workers and the dacctl communicate using etcd: http://etcd.io

The `dacd` service makes use of Ansible roles (./fs-ansible) to create the Lustre
or BeeGFS filesystems on demand, using the NVMe drives that have been assigned
by the data accellerator. Mounting on the compute nodes is currently done via ssh
(as the user running dacd), rather than using Ansible.

### Slurm Integration

When you request a burst buffer via Slurm, the Cray DataWarp burst buffer plugin is used
to communicate to `dacctl`. We support both per job and persistent burst buffers.

You can create a peristent burst buffer by submitting a job like this:
```
#BB create_persistent name=mytestbuffer capacity=4000GB access=striped type=scratch
```

To use the above buffer in a job, add the following pragma in a job submission script:
```
#DW persistentdw name=mytestbuffer
#DW jobdw capacity=2TB access_mode=striped,private type=scratch
#DW swap 1GB
#DW stage_in source=~/mytestinputfile destination=\$DW_JOB_STRIPED/filename1 type=file
#DW stage_out source=\$DW_JOB_STRIPED/outdir destination=~/test7outputdir type=directory
```

Please note the above job does the following:

* mounts the persistent buffer called mytestbuffer
* create a per job buffer of 2TB in size, with extra space requested for the swap
* mounts a shared directory on every compute node
* also mounts a private directory that is specific to each compute node
* adds a 1GB swap file on each compute node
* before the job starts copies in the specified file
* after the job completes copied out the specified output file

To delete a persistent buffer you submit the following:
```
#BB destroy_persistent name=mytestbuffer
```

Further details on the Slurm intergration can be found here:
https://slurm.schedmd.com/burst_buffer.html

### Orchestrator golang Code Tour

The golang code is built using make, including creating a tarball that includes
all the ansible that needs to be installed on all the dacd nodes. Currently we
use CircleCI to run the unit tests on every pull request before it is merged
into master, this includes generating tarballs for all commits.

The following tests are currently expected to work:

* unit tests (make tests)
* Slurm integration tests using Docker compose (see below on how to run ./docker-slurm)
* Full end to end test deployment using ansible to install systemd unit files, with SSL certs for etcd, aimed at testing the Ansible inside virtual machines (./dac-ansible)

The following tests are currently a work in progress:

* functional tests for etcd (make test-func runs dac-func-test golang binary)

### Packages

* "github.com/RSE-Cambridge/data-acc/internal/pkg/registry" is the core data model of the PoolRegistry and VolumeRegistry

* "github.com/RSE-Cambridge/data-acc/internal/pkg/keystoreregistry" depends on a keystore interface, and implements
  the PoolRegistry and VolumeRegistry

* "github.com/RSE-Cambridge/data-acc/internal/pkg/etcdregistry" implements the keystore interface using etcd

* "github.com/RSE-Cambridge/data-acc/internal/pkg/lifecycle" provides business logic on top of registry interface

* "github.com/RSE-Cambridge/data-acc/internal/pkg/pfsprovider" provides a plugin interface, and various implementations
  that implement needed configuration and setup of the data accelerator node

* "github.com/RSE-Cambridge/data-acc/internal/pkg/dacctl" this does the main work of implementing the CLI tool.
  While we use "github.com/urfave/cli" is used to build the cli, we keep this at arms length via a CliContext interface.

* "github.com/RSE-Cambridge/data-acc/internal/pkg/fileio" interfaces to help with unit testing file reading and writing

* "github.com/RSE-Cambridge/data-acc/internal/pkg/mocks" these are mock interfaces needed for unit testing, created
  using "github.com/golang/mock/gomock" and can be refreshed by running a [build script](build/rebuild_mocks.sh).

## Golang Build and Test (using make)

Built with golang 1.14.x, using go modules for dependency management.

gomock v1.1.1 is used to generate mocks. The version is fixed to stop
conflicts with etcd 3.3.x.

To build all the golang code and run unit tests locally:
```
cd ~/go/src/github.com/RSE-Cambridge/data-acc
make
make test
```

To build the tarball:
```
make tar
```

There is an experimental effort to build things inside a docker container here:
```
make docker
```

To mimic what is happening in circleci locally please see:
https://circleci.com/docs/2.0/local-cli/

## License

This work is licensed under the Apache 2.
Please see LICENSE file for more information.

Copyright © 2018-2019 Alasdair James King, University of Cambridge

Copyright © 2018-2020 John Garbutt, StackHPC Ltd
