#!/bin/bash

set -eu -o pipefail

make
#make test

export ETCDCTL_ENDPOINTS=127.0.0.1:2379
export ETCDCTL_API=3

echo
echo run tests:
## if etcd is not running in docker or localy run etcd &

if [ ! -f ./bin/amd64/dacd ]; then
  $GOPATH/bin/dacd &
  $GOPATH/bin/etcd-keystore-func-test
  kill %1

  echo
  echo see side effects:
  etcdctl get --prefix ""
else
  ./bin/amd64/dacd &
  ./bin/amd64/etcd-keystore-func-test
  kill %1

  echo
  echo see side effects:
  etcdctl get --prefix ""
fi
