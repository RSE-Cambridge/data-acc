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

"""
Generic API to access the burst buffer from an execution engine.
Involves creating, destroying, staging data in and staging data out, etc.
"""

from burstbuffer import model
from burstbuffer.provision import api as provision
from burstbuffer.registry import api as registry

TB_IN_BYTES = 1 * 10 ** 12
GiB_IN_BYTES = 1073741824


def get_all_pool_stats():
    # TODO(johngarbutt) Do we need to model the default pool,
    # or is that something for Slurm configuration?
    return [
        model.PoolStats("dedicated_nvme",
                        total_slices=20, free_slices=10,
                        slice_bytes=TB_IN_BYTES),
    ]


def get_all_buffers():
    return [
        model.Buffer(1, 1001, "dedicated_nvme", 2, 2 * 10 ** 12, 42),
        model.Buffer(2, 1001, "dedicated_nvme", 4, 4 * 10 ** 12,
                     persistent=True, name="testpersistent"),
    ]


def add_buffer(buff_request):
    id = buff_request.job_id
    if not id:
        id = buff_request.name
    buff = registry.add_new_buffer(id, buff_request)
    assignments = provision.assigned_slices(id)
    # TODO(johngarbutt) need to wait for mountpoint, etc?
    print(assignments)
    return buff


def delete_buffer(buffer_id):
    try:
        provision.unassign_slices(buffer_id)
        registry.delete_buffer(buffer_id)
    except Exception as e:
        print("Ignoring delete error: %" % e)
