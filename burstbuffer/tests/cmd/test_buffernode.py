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

import fixtures
import mock
import socket
import testtools

from burstbuffer.cmd import buffernode
from burstbuffer.provision import api


class TestBuffernode(testtools.TestCase):

    def setUp(self):
        super(TestBuffernode, self).setUp()
        stdout = self.useFixture(fixtures.StringStream('stdout')).stream
        self.useFixture(fixtures.MonkeyPatch('sys.stdout', stdout))
        stderr = self.useFixture(fixtures.StringStream('stderr')).stream
        self.useFixture(fixtures.MonkeyPatch('sys.stderr', stderr))

    @mock.patch.object(api, "startup")
    @mock.patch.object(socket, "gethostname")
    def test_startup(self, mock_hostname, mock_startup):
        mock_hostname.return_value = "test"

        result = buffernode.main(["startup"])

        self.assertEqual(0, result)
        mock_startup.assert_called_once_with("test")

    @mock.patch.object(api, "event")
    @mock.patch.object(socket, "gethostname")
    def test_event(self, mock_hostname, mock_event):
        mock_hostname.return_value = "test"

        result = buffernode.main(["event"])

        self.assertEqual(0, result)
        mock_event.assert_called_once_with("test")
