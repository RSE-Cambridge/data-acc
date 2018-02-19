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

import mock
import os
import subprocess

import testtools

from burstbuffer.registry import api


class TestAPI(testtools.TestCase):
    @mock.patch.object(subprocess, "check_output")
    def test_get_all_with_prefix(self, mock_cmd):
        mock_cmd.return_value = b"""
{"header":
    {"cluster_id":10316109323310759371,
     "member_id":15168875803774599630,
     "revision":1403,
     "raft_term":5},
    "kvs":[
      {"key":"aGVsbG8=","create_revision":1399,"mod_revision":1399,
       "version":1,"value":"dGVzdA=="},
      {"key":"aGVscG8=","create_revision":1399,"mod_revision":1399,
       "version":1,"value":"dGVzeA=="}
     ],"count":1}
"""
        with mock.patch.dict(os.environ, {"foo": "bar"}):
            result = api._get_all_with_prefix("hel")

        self.assertEqual("test", result["hello"])
        self.assertEqual("tesx", result["helpo"])
        self.assertEqual([
            'etcdctl', '--endpoints=http://localhost:2379', '-w',
            'json', 'get', '--prefix', 'hel'],
            mock_cmd.call_args[0][0])
