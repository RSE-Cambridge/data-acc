# Copyright 2016 The Kubernetes Authors.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

# The binary to build (just the basename).
BIN := fakewarp

# This repo's root import path (under GOPATH).
export PKG := github.com/RSE-Cambridge/data-acc

# This version-strategy uses git tags to set the version string
VERSION := $(shell git describe --tags --always --dirty)
#
# This version-strategy uses a manual value to set the version string
export CGO_ENABLED=0
export GOARCH=amd64

all: build

build-%:
	@$(MAKE) --no-print-directory ARCH=$* build

all-build: $(addprefix build-, $(ALL_ARCH))

build: build-dirs
	@go install -x -installsuffix "statiic" -ldflags "-X ${PKG}/pkg/version.VERSION=${VERSION}" ./...

test: build-dirs
	/bin/sh -c " ./build/test.sh $(SRC_DIRS)"

build-dirs:
	@mkdir -p bin/$(ARCH)
	@mkdir -p .go/src/$(PKG) .go/pkg .go/bin .go/std/$(ARCH)

clean: bin-clean

bin-clean:
	rm -rf .go bin
