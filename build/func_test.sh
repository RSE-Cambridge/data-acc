#!/bin/bash

set -eu -o pipefail

DOWNLOAD_URL=https://github.com/etcd-io/etcd/releases/download
ETCD_VER=v3.3.10
ETCDCTL_API=3

if [ ! -d /tmp/etcd-download ]; then
    curl -L ${DOWNLOAD_URL}/${ETCD_VER}/etcd-${ETCD_VER}-linux-amd64.tar.gz -o /tmp/etcd-${ETCD_VER}-linux-amd64.tar.gz
    rm -rf /tmp/etcd-download
    mkdir /tmp/etcd-download
    tar xzvf /tmp/etcd-${ETCD_VER}-linux-amd64.tar.gz -C /tmp/etcd-download --strip-components=1
    rm -f /tmp/etcd-${ETCD_VER}-linux-amd64.tar.gz
fi
/tmp/etcd-download/etcd --version
ETCDCTL_API=3 /tmp/etcd-download/etcdctl version

make

# be sure to kill etcd even if we hit an error
trap 'kill $(jobs -p)' EXIT
/tmp/etcd-download/etcd &

ETCDCTL_ENDPOINTS=127.0.0.1:2379 ./bin/dac-func-test
