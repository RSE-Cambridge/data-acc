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

import time


class PoolStats(object):
    """Pool of buffer storage."""
    def __init__(self, name, total_slices, free_slices, slice_bytes):
        self.name = name
        self.total_slices = total_slices
        self.free_slices = free_slices
        self.slice_bytes = slice_bytes


class Buffer(object):
    """Buffer is an assignment of io_slices"""
    def __init__(self, id, user_id,
                 pool_name, capacity_slices, capacity_bytes,
                 job_id=None, name=None, persistent=False):
        self.created_at = int(time.time())
        self.id = id
        self.user_id = user_id

        self.pool_name = pool_name
        self.capacity_slices = capacity_slices
        self.capacity_bytes = capacity_bytes  # TODO(johng) redundant data

        self.job_id = job_id
        self.name = name
        self.persistent = persistent
