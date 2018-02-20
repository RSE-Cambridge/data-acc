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

import os

from burstbuffer.registry import api

ASSIGNED_SLICES_KEY = "bufferhosts/assigned_slices/%s"
ALL_SLICES_KEY = "bufferhosts/all_slices/%s/%s"

FAKE_DEVICE_COUNT = 12
FAKE_DEVICE_ADDRESS = "nvme%sn1"
FAKE_DEVICE_SIZE_BYTES = int(1.5 * 2 ** 40)  # 1.5 TB


def _get_all_slices():
    fake_devices = []
    for i in range(FAKE_DEVICE_COUNT):
        fake_devices.append(FAKE_DEVICE_ADDRESS % i)
    return fake_devices


def _refresh_slices(hostname, slices):
    for s in slices:
        # TODO(johngarbut) check verison for creations, etc
        key = ALL_SLICES_KEY % (hostname, s)
        api._etcdctl("put '%s' '%s'" % (key, FAKE_DEVICE_SIZE_BYTES))


def _get_assigned_slices(hostname):
    prefix = ASSIGNED_SLICES_KEY % hostname
    raw_assignments = api._get_all_with_prefix(prefix)
    current_devices = _get_all_slices()

    assignemnts = {}
    for key in raw_assignments:
        device = key[(len(prefix) + 1):]
        if device not in current_devices:
            raise Exception("assignment to unknown device %s!!" % device)
        assignemnts[device] = raw_assignments[key]
    return assignemnts


def startup(hostname):
    all_slices = _get_all_slices()
    _refresh_slices(hostname, all_slices)

    return _get_assigned_slices(hostname)


def event(hostname):
    for key in os.environ:
        if key.startswith("ETCD_WATCH_"):
            short_key = key[len("ETCD_WATCH_"):]
            print("%s: %s" % (short_key, os.environ[key]))
    return _get_assigned_slices(hostname)
