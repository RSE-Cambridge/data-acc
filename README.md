# Data Accelerator

[![CircleCI](https://circleci.com/gh/RSE-Cambridge/data-acc.svg?style=svg&circle-token=4042ee71fb486efc320ce64b7b568afd4f9e0b38)](https://circleci.com/gh/RSE-Cambridge/data-acc)

<!-- [![Build Status](https://travis-ci.org/RSE-Cambridge/data-acc.svg?branch=master)](https://travis-ci.org/RSE-Cambridge/data-acc)
[![Go Report Card](https://goreportcard.com/badge/github.com/RSE-Cambridge/data-acc)](https://goreportcard.com/report/github.com/RSE-Cambridge/data-acc)
[![Godoc](http://img.shields.io/badge/go-documentation-blue.svg?style=flat-square)](https://godoc.org/github.com/RSE-Cambridge/data-acc)
[![Releases](https://img.shields.io/github/release/RSE-Cambridge/data-acc/all.svg?style=flat-square)](https://github.com/JohnGarbutt/RSE-Cambridge/data-acc)
[![LICENSE](https://img.shields.io/github/license/RSE-Cambridge/data-acc.svg?style=flat-square)](https://github.com/RSE-Cambridge/data-acc/blob/master/LICENSE)
-->

Data Accelerator uses commodity storage to accelerate HPC jobs.
Currently targeting initial integration with the Slurm Burst Buffer plugin,
with Lustre over Intel P4600 attached to Dell R730 with 2x100Gb/s OPA.

When you request a burst buffer via Slurm, the Cray Data Warp plugin is used
to communicate to dacctl to orchestrate the creation of the burst buffer via
the data accelerator. The user requests a certain capacity, which is rounded
up to a number of NVMe devices. The choen parallel filesystem is created that
exposes those NVMe devices to the compute nodes Slurm chooses for the Slurm
job that requested the burst buffer.

## Code Guided Tour

There are two key binaries produced by the golang based code:

* [dacd](cmd/dacd): service that runs on the storage nodes to orchestrate filesystem creation
* [dacctl](cmd/dacctl): CLI tool used by Slurm Cray DataWarp burst buffer plugin to orchestration burst buffer creation

All the dacd workers and the dacctl communicate using etcd: http://etcd.io

The dacd service makes use of Ansible roles (./fs-ansible) to create the Lustre
or BeeGFS filesystems on demand, using the NVMe drives that have been assigned
by the data accellerator. Mounting on the compute nodes is done via ssh
(as the user running dacd), rather than using Ansible.

The golang code is built using make, including creating a tarball that includes
all the ansible that needs to be installed on all the dacd nodes. Currently we
use CircleCI to run the unit tests on every pull request before it is merged
into master, this includes generating tarballs for all commits.

The following tests are currently expected to work:

* unit tests (make tests)
* Slurm integration tests using Docker compose (see below on how to run ./docker-slurm)

The following tests are currently a work in progress:

* full end to end test deployment using ansible to install systemd unit files, with SSL certs for etcd, aimed at testing the Ansible inside virtual machines (./dac-ansible)
* older more broken end to end deployment using docker-compose (./ansible)
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

## Docker Compose based Integration Test

To see end to end demo with Slurm 18.08
(but without running fs-ansible and not ssh-ing to compute nodes to mount):
```
cd docker-slurm
./update_burstbuffer.sh
```

To clean up after the demo, including removing all docker volumes:
```
docker-compose down --vol
```

## Golang Build and Test (using make)

Ensure you checkout this code in a working golang 1.11 workspace, including setting $GOPATH as required:
https://golang.org/doc/install#testing

dep 0.5.0 is used to manage dependencies. To install dep please read:
https://golang.github.io/dep/docs/installation.html#binary-installation

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
