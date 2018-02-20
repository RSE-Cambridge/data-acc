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
import testtools

from burstbuffer.provision import api


class TestProvisionAPI(testtools.TestCase):
    @mock.patch.object(api, "_get_assigned_slices")
    @mock.patch.object(api, "_refresh_slices")
    @mock.patch.object(api, "_get_local_hardware")
    def test_statup(self, mock_hardware, mock_refresh, mock_assigned):
        mock_hardware.return_value = "fake_hardware"
        mock_assigned.return_value = "fake"

        result = api.startup("test")

        self.assertEqual("fake", result)
        mock_hardware.assert_called_once_with()
        mock_refresh.assert_called_once_with("test", "fake_hardware")
        mock_assigned.assert_called_once_with("test")

    @mock.patch.object(api, "_get_assigned_slices")
    @mock.patch.object(api, "_get_event_info")
    def test_event(self, mock_einfo, mock_assigned):
        mock_assigned.return_value = "fake"

        result = api.event("test")

        self.assertEqual("fake", result)
        mock_einfo.assert_called_once_with()

    @mock.patch.object(api, "_get_env")
    def test_get_event_info_fails(self, mock_env):
        mock_env.return_value = {}

        self.assertRaises(KeyError, api._get_event_info)

    @mock.patch.object(api, "_get_env")
    def test_get_event_info(self, mock_env):
        mock_env.return_value = dict(
            ETCD_WATCH_EVENT_TYPE="event_type",
            ETCD_WATCH_REVISION="revision",
            ETCD_WATCH_KEY="key",
            ETCD_WATCH_VALUE="value")

        result = api._get_event_info()

        expected = dict(
            event_type="event_type",
            key="key",
            revision="revision",
            value="value")
        self.assertDictEqual(expected, result)
