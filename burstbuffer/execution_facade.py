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

from burstbuffer import model

TB_IN_BYTES = 1 * 10 ** 12
GiB_IN_BYTES = 1073741824


def get_all_pool_stats():
    return [
        model.PoolStats("dedicated_nvme",
                        total_slices=20, free_slices=10,
                        slice_bytes=TB_IN_BYTES),
        model.PoolStats("test_pool",  # "dedicated_nvme",
                        total_slices=2048, free_slices=2046,
                        slice_bytes=GiB_IN_BYTES),
    ]
