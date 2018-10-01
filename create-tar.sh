#!/bin/bash

set -eux

cd ..
tar -cvzf ./data-acc/data-acc.tgz ./data-acc/bin/amd64/data-acc-brick-host ./data-acc/bin/amd64/fakewarp ./data-acc/fs-ansible
sha256sum ./data-acc/data-acc.tgz
