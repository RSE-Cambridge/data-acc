#!/bin/bash

set -eux

cd ..
tar -cvzf ./data-acc/data-acc.tgz ./data-acc/bin/amd64/data-acc-brick-host ./data-acc/bin/amd64/dacctl ./data-acc/fs-ansible ./data-acc/tools/*.sh
sha256sum ./data-acc/data-acc.tgz
