#!/usr/bin/python3
# Licensed under the Apache License, Version 2.0 (the "License"); you may
# not use this file except in compliance with the License. You may obtain
# a copy of the License at
#
#      http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS, WITHOUT
# WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the
# License for the specific language governing permissions and limitations
# under the License.

import base64
import json
import os
import shlex
import subprocess

ETCD_ENDPOINTS = "http://localhost:2379"


def _etcdctl(cmd, parse_json=True):
    cmd = "etcdctl --endpoints=%s -w json %s" % (ETCD_ENDPOINTS, cmd)
    split = shlex.split(cmd)
    env = dict(os.environ, ETCDCTL_API="3")

    raw = subprocess.check_output(split, env=env).decode("utf-8")

    if parse_json:
        return json.loads(raw)
    return raw


def _get(args):
    result = _etcdctl("get %s" % args)
    kvs = result.get('kvs', [])

    result = {}
    for key_value in kvs:
        key = base64.b64decode(key_value["key"]).decode("utf-8")
        value = base64.b64decode(key_value["value"]).decode("utf-8")
        result[key] = value
    return result


def _get_all_with_prefix(prefix):
    return _get("--prefix %s" % prefix)


def add_new_buffer(buffer_id, buffer_info):
    key = "buffers/%s" % buffer_id
    value = json.dumps(buffer_info)

    put_result = _etcdctl("put '%s' '%s'" % (key, value))
    revision = put_result['header']['revision']

    get_result = _etcdctl("get %s" % key).get('kvs')
    version = None
    if len(get_result) == 1:
        version = get_result[0]['version']
    if version != 1:
        # TODO move to txn...
        raise Exception("buffer already created")

    return revision


def delete_buffer(buffer_id):
    key = "buffers/%s" % buffer_id
    del_result = _etcdctl("del '%s'" % key)
    keys_deleted = del_result.get('deleted', 0)
    if keys_deleted == 0:
        raise Exception("Buffer already deleted")
    if keys_deleted > 1:
        raise Exception("WARNING: deleted too many buffers!")
    return del_result['header']['revision']


def list_buffers():
    return _get_all_with_prefix(prefix="buffers/")


if __name__ == '__main__':
    print(add_new_buffer("test", {"persistent": True}))
    print(list_buffers())
    print(delete_buffer("test"))
