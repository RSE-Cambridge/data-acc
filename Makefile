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

export GO111MODULE=on
VERSION := $(shell git describe --tags --always --dirty)

all: buildlocal format buildmocks test

buildlocal:
	mkdir -p `pwd`/bin
	GOBIN=`pwd`/bin go install -mod=vendor -ldflags "-X github.com/RSE-Cambridge/data-acc/pkg/version.VERSION=${VERSION}" -v ./...
	ls -l `pwd`/bin

format:
	go fmt ./...

buildmocks:
	./build/rebuild_mocks.sh

test: 
	mkdir -p `pwd`/bin
	go vet ./...
	go test -mod=vendor -cover -race -coverprofile=./bin/coverage.txt ./...

test-func:
	./build/func_test.sh

clean:
	go clean ./...
	rm -rf `pwd`/bin
	rm -rf /tmp/etcd-download

tar: clean buildlocal
	tar -cvzf ./bin/data-acc-`git describe --tag --dirty`.tgz ./bin/dacd ./bin/dacctl ./fs-ansible ./tools/*.sh
	sha256sum ./bin/data-acc-`git describe --tag --dirty`.tgz > ./bin/data-acc-`git describe --tag`.tgz.sha256
	go version > ./bin/data-acc-`git describe --tag`.go-version
