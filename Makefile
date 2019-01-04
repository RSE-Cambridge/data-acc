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


all: deps buildlocal format test

buildlocal:
	mkdir -p `pwd`/bin
	GOBIN=`pwd`/bin go install -v ./...
	ls -l `pwd`/bin

format:
	go fmt ./...

test: 
	mkdir -p `pwd`/bin
	./build/rebuild_mocks.sh
	go vet ./...
	go test -cover -race -coverprofile=./bin/coverage.txt ./...

test-func:
	./build/func_test.sh

clean:
	go clean
	rm -rf `pwd`/bin
	rm -rf /tmp/etcd-download

deps:
	dep ensure

tar: clean buildlocal
	tar -cvzf ./bin/data-acc-`git describe --tag --dirty`.tgz ./bin/dacd ./bin/dacctl ./fs-ansible ./tools/*.sh
	sha256sum ./bin/data-acc-`git describe --tag --dirty`.tgz > ./bin/data-acc-`git describe --tag`.tgz.sha256

dockercmd=docker run --rm -it -v ~/go:/go -w /go/src/github.com/RSE-Cambridge/data-acc golang:1.11

docker:
	$(dockercmd) go install -v ./...

installdep:
	curl https://raw.githubusercontent.com/golang/dep/master/install.sh | sh
