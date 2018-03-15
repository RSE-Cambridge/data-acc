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

import paramiko

BRICK_PATH = "/data/glusterfs/%s/brick"


def _exec_command(command, hostname, port=22, username="root"):
    client = paramiko.SSHClient()
    client.load_system_host_keys()
    client.set_missing_host_key_policy(paramiko.AutoAddPolicy())
    client.connect('ssh.example.com', port, username, timeout=10)

    stdin, stdout, stderr = client.exec_command('ls -l', timeout=60.0)

    client.close()

    return (stdin, stdout, stderr)


def _create_bricks(slices):
    bricks = []
    for io_slice in slices:
        host, device = io_slice.split(":")
        path = BRICK_PATH % device
        _exec_command("mkdir %s" % path, host)
        bricks.append(path)
    return bricks


def _volume_create(gluster_host, volume_name, *bricks):
    create_base = "gluster volume create %s %s"
    command = create_base % volume_name, " ".join(bricks)
    _exec_command(command, gluster_host)


def _volume_start(gluster_host, volume_name):
    start_base = "gluster volume start %s"
    _exec_command(start_base % volume_name, gluster_host)


def _volume_prefer_local(gluster_host, volume_name):
    prefer_base = "gluster volume set %s cluster.nufa enable on"
    _exec_command(prefer_base % volume_name, gluster_host)


def _volume_stop(gluster_host, volume_name):
    stop_base = "gluster volume stop %s"
    _exec_command(stop_base % volume_name, gluster_host)


def _volume_delete(gluster_host, volume_name):
    stop_base = "gluster volume delete %s"
    _exec_command(stop_base % volume_name, gluster_host)


def setup_volume(gluster_host, volume_name, *slices):
    bricks = _create_bricks(gluster_host, slices)
    _volume_create(gluster_host, volume_name, *bricks)
    _volume_start(gluster_host, volume_name)


def remove_volume(gluster_host, volume_name):
    # TODO(johngarbutt) might be cleaner to clean all bricks here?
    _volume_stop(gluster_host, volume_name)
    _volume_delete(gluster_host, volume_name)


def clean_brick(gluster_host, device_name):
    path = BRICK_PATH % device_name
    _exec_command("rm -rf %s" % path)
