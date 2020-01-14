#!/bin/bash

source "$( dirname $0 )/tests/config"

for i in tests/*.sh; do
    timeout $timeout $i || echo test $i timed out
done
