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
import time

import testtools

from burstbuffer import execution_facade
from burstbuffer import model
from burstbuffer.provision import api as provision
from burstbuffer.registry import api as registry


class TestExecutionFacade(testtools.TestCase):

    def test_get_all_pool_stats(self):
        all_pool_stats = execution_facade.get_all_pool_stats()

        self.assertEqual(1, len(all_pool_stats))
        self.assertEqual("dedicated_nvme", all_pool_stats[0].name)
        self.assertEqual(20, all_pool_stats[0].total_slices)
        self.assertEqual(10, all_pool_stats[0].free_slices)
        self.assertEqual(10 ** 12, all_pool_stats[0].slice_bytes)

    @mock.patch.object(registry, "list_buffers")
    def test_get_all_buffers(self, mock_list):
        mock_list.return_value = "fake"

        result = execution_facade.get_all_buffers()

        self.assertEqual("fake", result)

    @mock.patch.object(time, "time")
    @mock.patch.object(provision, "assign_slices")
    @mock.patch.object(registry, "add_new_buffer")
    def test_add_buffer_with_jobid(self, mock_add, mock_assign, mock_time):
        mock_time.return_value = 1519172799
        buff_request = model.Buffer(
            None, 1001, "dedicated_nvme", 2, 2 * 10 ** 12, 42)

        execution_facade.add_buffer(buff_request)

        mock_add.assert_called_once_with(42, {
            'pool_name': 'dedicated_nvme', 'created_at': 1519172799,
            'capacity_slices': 2, 'capacity_bytes': 2000000000000,
            'job_id': 42, 'user_id': 1001, 'user_agent': None,
            'name': None, 'id': None, 'persistent': False})
        mock_assign.assert_called_once_with(42)

    @mock.patch.object(provision, "unassign_slices")
    @mock.patch.object(registry, "delete_buffer")
    def test_delete_buffer(self, mock_delete, mock_unassign):
        mock_delete.side_effect = Exception
        mock_unassign.side_effect = Exception

        execution_facade.delete_buffer("bufferid")

        mock_delete.assert_called_once_with("bufferid")
        mock_unassign.assert_called_once_with("bufferid")
