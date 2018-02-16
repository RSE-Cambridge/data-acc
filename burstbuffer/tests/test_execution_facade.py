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


class TestExecutionFacade(testtools.TestCase):

    def test_get_all_pool_stats(self):
        all_pool_stats = execution_facade.get_all_pool_stats()

        self.assertEqual(1, len(all_pool_stats))
        self.assertEqual("dedicated_nvme", all_pool_stats[0].name)
        self.assertEqual(20, all_pool_stats[0].total_slices)
        self.assertEqual(10, all_pool_stats[0].free_slices)
        self.assertEqual(10 ** 12, all_pool_stats[0].slice_bytes)

    @mock.patch.object(time, "time")
    def test_get_all_buffers(self, mock_time):
        mock_time.return_value = 123.45

        result = execution_facade.get_all_buffers()

        self.assertEqual(2, len(result))

        self.assertEqual(1, result[0].id)
        self.assertEqual(1001, result[0].user_id)
        self.assertEqual(42, result[0].job_id)
        self.assertEqual(2000000000000, result[0].capacity_bytes)
        self.assertEqual(2, result[0].capacity_slices)
        self.assertEqual(123, result[0].created_at)
        self.assertEqual("dedicated_nvme", result[0].pool_name)
        self.assertIsNone(result[0].name)
        self.assertFalse(result[0].persistent)

        self.assertEqual(2, result[1].id)
        self.assertTrue(result[1].persistent)
        self.assertEqual(4, result[1].capacity_slices)
