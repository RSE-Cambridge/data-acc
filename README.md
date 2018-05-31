# Data Accelerator

<!-- [![Build Status](https://travis-ci.org/JohnGarbutt/pfsaccel.svg?branch=master)](https://travis-ci.org/JohnGarbutt/pfsaccel)
[![Go Report Card](https://goreportcard.com/badge/github.com/johngarbutt/pfsaccel)](https://goreportcard.com/report/github.com/johngarbutt/pfsaccel)
[![Godoc](http://img.shields.io/badge/go-documentation-blue.svg?style=flat-square)](https://godoc.org/github.com/JohnGarbutt/pfsaccel)
[![Releases](https://img.shields.io/github/release/JohnGarbutt/pfsaccel/all.svg?style=flat-square)](https://github.com/JohnGarbutt/pfsaccel/releases)
[![LICENSE](https://img.shields.io/github/license/JohnGarbutt/pfsaccel.svg?style=flat-square)](https://github.com/JohnGarbutt/pfsaccel/blob/master/LICENSE)
-->

Data Accelerator uses commodity storage to accelerate HPC jobs.
Currently targeting initial integration with Slurm Burst Buffer plugin,
with Lustre over Intel P4600 attached to Dell R730 with 2x100Gb/s OPA.

To try this out, run etcd then fetch the functional test via:
```
go get https://github.com/RSE-Cambridge/cmd/data-acc-func-test
```

To build it locally:
```
make
```
