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

    def test_get_instances(self):
        result = fakewarp_facade.get_instances()

        instances = result['instances']
        self.assertEqual(2, len(instances))
        self.assertEqual(1, instances[0]['id'])
        self.assertEqual(2, instances[0]['capacity']['nodes'])
        self.assertEqual(2000000000000, instances[0]['capacity']['bytes'])
        self.assertEqual(1, instances[0]['links']['session'])

        self.assertEqual(2, instances[1]['id'])
        self.assertEqual(4, instances[1]['capacity']['nodes'])
