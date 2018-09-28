#!/bin/bash

set -eux

cd bin/amd64/
tar -cvzf ../../data-acc.tgz ./data-acc-brick-host ./fakewarp
