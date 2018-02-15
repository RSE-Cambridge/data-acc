# Generic BurstBuffer

This project is designed to orchestrate the creation of a burst
buffer built using commodity hardware. The buffer can be accessed
by compute nodes using existing parallel file systems.

Users are expected to request a number of I/O slots, which is a
fixed amount of capacity associated with a fixed amount of
bandwidth. The aim being each burst buffer is independent of any
other burst buffer in the system.

Initially we are targeting two access modes:

1. Shared Scratch (all compute nodes access the same namespace,
   files are striped across all burst buffer nodes using PFS)
2. Dedicated Scratch (accessed by a single node, files striped
   across all burst buffer nodes)

In addition to orchestrating the creation of the buffer, there is
also an option to stage in and stage out files to and from an
existing slower storage tier. For example you might copy from
spinning disk based Lustre to an NVMe backed burst buffer.

## Components

Currently we are targeting integrating with Slurm 17.02.9

The packing is based on:
https://github.com/giovtorres/slurm-docker-cluster

## Installation

```Console
virtualenv .venv
. .venv/bin/activate
pip install -U pip
pip install .
```

## Usage

Try out the fake pools example:
```Console
fakewarp --fuction pools
```

## Running tests

To run unit tests:

```Console
tox -epy35
```

To check the style:

```Console
tox -epep8
```
