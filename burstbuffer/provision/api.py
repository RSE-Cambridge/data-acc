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

import collections
import os
import random

from burstbuffer.registry import api as registry

ASSIGNED_SLICES_PREFIX = "bufferhosts/assigned_slices/"
ASSIGNED_SLICES_KEY = ASSIGNED_SLICES_PREFIX + "%s"
ALL_SLICES_PREFIX = "bufferhosts/all_slices/"
ALL_SLICES_KEY = ALL_SLICES_PREFIX + "%s/%s"

FAKE_DEVICE_COUNT = 12
FAKE_DEVICE_ADDRESS = "nvme%sn1"
FAKE_DEVICE_SIZE_BYTES = int(1.5 * 2 ** 40)  # 1.5 TB


class UnexpectedBufferAssignement(Exception):
    pass


class UnableToAssignSlices(Exception):
    pass


def _get_local_hardware():
    fake_devices = []
    for i in range(FAKE_DEVICE_COUNT):
        fake_devices.append(FAKE_DEVICE_ADDRESS % i)
    return fake_devices


def _update_data(data, ensure_first_version=False):
    # TODO(johngarbutt) should be done in a single transaction
    for key, value in data.items():
        # TODO(johngarbutt) check version is 0 when ensure_first_version
        registry._etcdctl("put '%s' '%s'" % (key, value))


def _refresh_slices(hostname, hardware):
    slices_info = {}
    for device in hardware:
        key = ALL_SLICES_KEY % (hostname, device)
        slices_info[key] = FAKE_DEVICE_SIZE_BYTES
    _update_data(slices_info)


def _get_assigned_slices(hostname):
    prefix = ASSIGNED_SLICES_KEY % hostname
    raw_assignments = registry._get_all_with_prefix(prefix)
    current_devices = _get_local_hardware()

    assignments = {}
    for key in raw_assignments:
        device = key[(len(prefix) + 1):]
        if device not in current_devices:
            raise UnexpectedBufferAssignement(device)
        assignments[device] = raw_assignments[key]
    return assignments


def startup(hostname):
    all_slices = _get_local_hardware()
    _refresh_slices(hostname, all_slices)

    return _get_assigned_slices(hostname)


def _get_env():
    return os.environ


def _get_event_info():
    env = _get_env()

    event_type = env["ETCD_WATCH_EVENT_TYPE"].strip('"')
    revision = env["ETCD_WATCH_REVISION"].strip('"')
    key = env["ETCD_WATCH_KEY"].strip('"')
    value = env['ETCD_WATCH_VALUE'].strip('"')

    return dict(
        event_type=event_type,
        revision=revision,
        key=key,
        value=value)


def event(hostname):
    event_info = _get_event_info()
    print(event_info)
    # TODO(johngarbutt) write key to say provision worked,
    # and write out fake mountpoint for slice 0

    if event_info['event_type'] == "PUT":
        device_name = event_info['key'].split('/')[-1]
        _, buffer_id, __, slice_id = event_info['value'].split('/')
        print("device %s for buffer %s slice number %s" % (
              device_name, buffer_id, slice_id))
        if int(slice_id) == 0:
            buffer_slices = _get_buffer_slices(buffer_id)
            slices = {}
            for buffer_key, slice_key in buffer_slices.items():
                slice_number = buffer_key.split('/')[-1]
                slice_info = slice_key.split('/')
                server = slice_info[-2]
                device = slice_info[-1]
                server = "gluster" + server[:-1]
                slices[slice_number] = "%s:%s" % (server, device)
            slice_list = " ".join(slices)
            print("ssh gluster1 gluster volume create %s %s" % (
                  buffer_id, slice_list))

    if event_info['event_type'] == "DELETE":
        device_name = event_info['key'].split('/')[-1]
        print("TODO: delete brick for %s" % device_name)
        # TODO(johngarbutt) volume deleted in buffer watcher maybe?

    return _get_assigned_slices(hostname)


def _get_all_slices():
    raw_slices = registry._get_all_with_prefix(ALL_SLICES_PREFIX)
    slices = []
    for key, _ in raw_slices.items():
        key_parts = key.split("/")
        host = key_parts[2]
        device = key_parts[3]
        slices.append((host, device))
    slices.sort()
    return slices


def _get_all_assigned_slices():
    raw_slices = registry._get_all_with_prefix(ASSIGNED_SLICES_PREFIX)
    slices = []
    for key, _ in raw_slices.items():
        key_parts = key.split("/")
        host = key_parts[2]
        device = key_parts[3]
        slices.append((host, device))
    slices.sort()
    return slices


def _get_available_slices_by_host():
    all_slices = _get_all_slices()
    all_assigned_slices = _get_all_assigned_slices()

    slices = collections.defaultdict(list)
    for host, device in all_slices:
        if (host, device) not in all_assigned_slices:
            slices[host].append(device)

    available = list([(host, device) for host, device in slices.items()])
    available.sort()
    return available


def _set_assignments(buffer_id, assignments):
    assignments = list(assignments)
    # stop 0 always being the same host
    random.shuffle(assignments)
    slice_data = {}
    buffer_data = {}

    for index in range(len(assignments)):
        host, device = assignments[index]

        # Add buffer to hosts slice assignments
        prefix = ASSIGNED_SLICES_KEY % host
        slice_key = "%s/%s" % (prefix, device)
        slice_value = "buffers/%s/slices/%s" % (buffer_id, index)
        slice_data[slice_key] = slice_value

        # Add slice host to buffer
        buffer_key = slice_value
        buffer_data[buffer_key] = slice_key

    # TODO(johngarbutt) ensure all updates were good in a transaction
    _update_data(buffer_data, ensure_first_version=True)
    # for now, ensure buffer written before slice events
    _update_data(slice_data, ensure_first_version=True)


def assign_slices(buffer_id):
    buffer_info = registry.get_buffer(buffer_id)
    required_slices = buffer_info['capacity_slices']

    avaliable_slices_by_host = _get_available_slices_by_host()
    if len(avaliable_slices_by_host) < required_slices:
        raise UnableToAssignSlices("Not enough hosts for %s" % required_slices)

    assignments = set()
    for host, devices in avaliable_slices_by_host:
        # avoid some contention by not just picking the first
        device = random.choice(devices)
        assignments.add((host, device))

    _set_assignments(buffer_id, assignments)

    return assignments


def _get_buffer_slices(buffer_id):
    return registry._get_all_with_prefix("buffers/%s/slices/" % buffer_id)


def _delete_all_keys(keys_to_delete):
    # Should be in a transaction
    for key in keys_to_delete:
        registry._etcdctl("del '%s'" % key)


def unassign_slices(buffer_id):
    slices = _get_buffer_slices(buffer_id)
    keys_to_delete = list(slices.values())
    keys_to_delete.sort()
    _delete_all_keys(keys_to_delete)
