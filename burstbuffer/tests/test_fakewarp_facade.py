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
from burstbuffer import fakewarp_facade
from burstbuffer import model


class TestFakewarpFacade(testtools.TestCase):

    @mock.patch.object(execution_facade, "get_all_pool_stats")
    def test_get_pools_with_no_pools(self, fake_get_all):
        fake_get_all.return_value = []

        result = fakewarp_facade.get_pools()

        self.assertEqual([], result['pools'])

    @mock.patch.object(execution_facade, "get_all_pool_stats")
    def test_get_pools_with_two_pools(self, fake_get_all):
        fake_get_all.return_value = [
            model.PoolStats("test1", 2, 0, 1),
            model.PoolStats("test2", 4, 1, 1),
        ]

        result = fakewarp_facade.get_pools()

        pools = result['pools']
        self.assertEqual(2, len(pools))
        self.assertEqual("test1", pools[0]["id"])
        self.assertEqual("test2", pools[1]["id"])

    @mock.patch.object(execution_facade, "get_all_buffers")
    def test_get_instances(self, mock_get):
        mock_get.return_value = [
            model.Buffer(1, 1001, "dedicated_nvme", 2, 2 * 10 ** 12, 42),
            model.Buffer(2, 1001, "dedicated_nvme", 4, 4 * 10 ** 12,
                         persistent=True, name="testpersistent"),
        ]

        result = fakewarp_facade.get_instances()

        instances = result['instances']
        self.assertEqual(2, len(instances))
        self.assertEqual(1, instances[0]['id'])
        self.assertEqual(2, instances[0]['capacity']['nodes'])
        self.assertEqual(2000000000000, instances[0]['capacity']['bytes'])
        self.assertEqual(1, instances[0]['links']['session'])

        self.assertEqual(2, instances[1]['id'])
        self.assertEqual(4, instances[1]['capacity']['nodes'])

    @mock.patch.object(execution_facade, "get_all_buffers")
    @mock.patch.object(time, "time")
    def test_get_sessions(self, mock_time, mock_get):
        mock_time.return_value = 123
        mock_get.return_value = [
            model.Buffer(1, 1001, "dedicated_nvme", 2, 2 * 10 ** 12, 42),
            model.Buffer(2, 1001, "dedicated_nvme", 4, 4 * 10 ** 12,
                         persistent=True, name="testpersistent"),
        ]

        result = fakewarp_facade.get_sessions()

        sessions = result['sessions']
        self.assertEqual(2, len(sessions))

        self.assertEqual('1', sessions[0]['id'])
        self.assertEqual(123, sessions[0]['created'])
        self.assertEqual(1001, sessions[0]['owner'])
        self.assertEqual("42", sessions[0]['token'])

        self.assertEqual('2', sessions[1]['id'])
        self.assertEqual(123, sessions[1]['created'])
        self.assertEqual(1001, sessions[1]['owner'])
        self.assertEqual("testpersistent", sessions[1]['token'])

    @mock.patch.object(time, "time")
    @mock.patch.object(execution_facade, "add_buffer")
    def test_setup_job_buffer(self, mock_add, mock_time):
        mock_time.return_value = 123
        mock_add.return_value = "fake"

        result = fakewarp_facade.setup_job_buffer(
            '13', 'SLURM', 'dwcache', '1GiB', 995)

        self.assertEqual("fake", result)
        expected = model.Buffer(
            '13', 995, 'dwcache', 2, 2 ** 30, persistent=False,
            job_id='13', user_agent='SLURM')

        mock_add.assert_called_once_with(expected)

    @mock.patch.object(execution_facade, "delete_buffer")
    def test_delete_buffer(self, mock_delete):
        fakewarp_facade.delete_buffer("buffertoken")

        mock_delete.assert_called_once_with("buffertoken")

    @mock.patch.object(time, "time")
    @mock.patch.object(execution_facade, "add_buffer")
    def test_add_persistent_buffer(self, mock_add, mock_time):
        mock_time.return_value = 123
        mock_add.return_value = "fake"

        result = fakewarp_facade.add_persistent_buffer(
            'alpha', 'SLURM', 'dwcache', '1GiB', 995,
            'striped', 'scratch')

        self.assertEqual("fake", result)
        expected = model.Buffer(
            'alpha', 995, 'dwcache', 2, 2 ** 30, persistent=True,
            name='alpha', user_agent='SLURM')

        mock_add.assert_called_once_with(expected)
