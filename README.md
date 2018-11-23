# Data Accelerator

<!-- [![Build Status](https://travis-ci.org/JohnGarbutt/pfsaccel.svg?branch=master)](https://travis-ci.org/JohnGarbutt/pfsaccel)
[![Go Report Card](https://goreportcard.com/badge/github.com/johngarbutt/pfsaccel)](https://goreportcard.com/report/github.com/johngarbutt/pfsaccel)
[![Godoc](http://img.shields.io/badge/go-documentation-blue.svg?style=flat-square)](https://godoc.org/github.com/JohnGarbutt/pfsaccel)
[![Releases](https://img.shields.io/github/release/JohnGarbutt/pfsaccel/all.svg?style=flat-square)](https://github.com/JohnGarbutt/pfsaccel/releases)
[![LICENSE](https://img.shields.io/github/license/JohnGarbutt/pfsaccel.svg?style=flat-square)](https://github.com/JohnGarbutt/pfsaccel/blob/master/LICENSE)
-->

Data Accelerator uses commodity storage to accelerate HPC jobs.
Currently targeting initial integration with the Slurm Burst Buffer plugin,
with Lustre over Intel P4600 attached to Dell R730 with 2x100Gb/s OPA.

## Demo (Docker)

To see end to end demo with Slurm (not currently working):
```
cd docker-slurm
./update_burstbuffer.sh
```

To clean up after the demo:
```
docker-compose down --vol
```

## Build

*Note* you can build this with or without docker, Use the Makefile.docker if you wish to use docker.

Ensure you have your GOPATH setup. Create a *go* directory containing a *src* and *bin* directory in your $HOME and add it to your $PATH.

To build it locally and run tests:
```
dep ensure -v 
make
make test
```

## Code Introduction

Here is a quick introduction to the code.

### Dependencies

All components depend on etcd (v3 API), both to store its state, and communicate via watching for state changes.

You can configure the system to access a locally running etcd by setting:
```
export ETCDCTL_ENDPOINTS=127.0.0.1:2379
```

Go's [dep](https://golang.github.io/dep/) is used to manage code dependencies. To get the dependencies sorted run:
```
dep ensure
```

To aid unit testing, generally all concrete implementations are created in the executable's main function.

### Executables

There is a [brick host manager](cmd/data-acc-brick-host), running on every data accelerator (DAC) node.
It is responsible for reporting the installed disks, and watching for volume updates relating to any volume
that has its primary brick (brick zero) assigned to that host.

There is [dacctl](cmd/dacctl) that helps integrate with Slurm's burst buffer system. It users the libs provided in
this repo to create jobs and volumes, including assigning bricks to volumes, and request and report on volume
state changes. The volume state changes are generally responded to by the above brick host manager.

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
