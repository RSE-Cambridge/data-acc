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
import random
import testtools

from burstbuffer.provision import api
from burstbuffer.registry import api as registry


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

    def test_get_local_hardware(self):
        result = api._get_local_hardware()

        self.assertEqual(12, len(result))
        self.assertEqual("nvme1n1", result[1])

    @mock.patch.object(registry, "_etcdctl")
    def test_update_data(self, mock_etcd):
        fake_data = dict(key1="value1", key2="value2")

        api._update_data(fake_data)

        mock_etcd.assert_has_calls([
            mock.call("put 'key1' 'value1'"),
            mock.call("put 'key2' 'value2'"),
        ], any_order=True)

    @mock.patch.object(api, "_update_data")
    def test_refresh_slices(self, mock_update):
        fake_hardware = ["1", "2"]

        api._refresh_slices("host", fake_hardware)

        expected = {
            'bufferhosts/all_slices/host/1': 1649267441664,
            'bufferhosts/all_slices/host/2': 1649267441664,
        }
        mock_update.assert_called_once_with(expected)

    @mock.patch.object(registry, "_get_all_with_prefix")
    def test_get_assigned_slices_fails(self, mock_get):
        mock_get.return_value = {"foo": "bar"}

        self.assertRaises(api.UnexpectedBufferAssignement,
                          api._get_assigned_slices, "host")

        mock_get.assert_called_once_with("bufferhosts/assigned_slices/host")

    @mock.patch.object(registry, "_get_all_with_prefix")
    def test_get_assigned_slices(self, mock_get):
        mock_get.return_value = {
            "bufferhosts/assigned_slices/host/nvme2n1": "buffers/fakename1",
            "bufferhosts/assigned_slices/host/nvme6n1": "buffers/fakename2"
        }

        result = api._get_assigned_slices("host")

        expected = {
            'nvme2n1': 'buffers/fakename1',
            'nvme6n1': 'buffers/fakename2',
        }
        self.assertDictEqual(expected, result)

    @mock.patch.object(api, "_set_assignments")
    @mock.patch.object(random, "choice")
    @mock.patch.object(api, "_get_available_slices_by_host")
    @mock.patch.object(registry, "get_buffer")
    def test_assign_slices(self, mock_get_buffer, mock_available, mock_choice,
                           mock_set):
        mock_choice.side_effect = lambda x: x[0]
        mock_get_buffer.return_value = {
            "id": "fakeid",
            "persistent": True,
            "capacity_slices": 3,
            "job_id": None,
            "name": "test",
        }
        mock_available.return_value = {
            "fakehost1": ["nvme0n1", "nvme0n2"],
            "fakehost2": ["nvme0n4", "nvme0n5"],
            "fakehost3": ["nvme0n1"],
        }

        api.assign_slices("fakeid")

        mock_get_buffer.assert_called_once_with("fakeid")
        expected_assignments = set([
            ('fakehost1', 'nvme0n1'),
            ('fakehost2', 'nvme0n4'),
            ('fakehost3', 'nvme0n1'),
        ])
        mock_set.assert_called_once_with("fakeid", expected_assignments)

    @mock.patch.object(random, "shuffle")
    @mock.patch.object(api, "_update_data")
    def test_set_assignments(self, mock_update, mock_shuffle):
        mock_shuffle.side_effect = lambda x: x.sort()
        assignments = set([
            ('fakehost1', 'nvme0n1'),
            ('fakehost2', 'nvme0n4'),
            ('fakehost3', 'nvme0n1'),
        ])

        api._set_assignments("fakeid", assignments)

        expected_data = {
            'bufferhosts/assigned_slices/fakehost1/nvme0n1':
                'buffers/fakeid/slices/0',
            'bufferhosts/assigned_slices/fakehost2/nvme0n4':
                'buffers/fakeid/slices/1',
            'bufferhosts/assigned_slices/fakehost3/nvme0n1':
                'buffers/fakeid/slices/2',
            'buffers/fakeid/slices/0':
                'bufferhosts/assigned_slices/fakehost1/nvme0n1',
            'buffers/fakeid/slices/1':
                'bufferhosts/assigned_slices/fakehost2/nvme0n4',
            'buffers/fakeid/slices/2':
                'bufferhosts/assigned_slices/fakehost3/nvme0n1',
        }
        mock_update.assert_called_once_with(
            expected_data, ensure_first_version=True)
