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

import traceback

import paramiko

BRICK_PATH = "/data/glusterfs/%s/brick"


def _exec_command(command, hostname, port=2222, username="root"):
    try:
        client = paramiko.SSHClient()
        client.load_system_host_keys()
        client.set_missing_host_key_policy(paramiko.AutoAddPolicy())
        print("connecting to: %s" % hostname)
        client.connect(hostname, port, username, timeout=30)

        print("running command: %s" % command)
        stdin, stdout, stderr = client.exec_command(command, timeout=60.0)

        stdout_lines = stdout.readlines()
        stderr_lines = stderr.readline()
        #print(stdout_lines)
        #print(stderr_lines)

        return (stdin, stdout, stderr)
    except Exception:
        print("uh oh")
        traceback.print_exc()
    finally:
        client.close()


def _create_bricks(slices):
    print(slices)
    bricks = []
    for io_slice in slices:
        host, device = io_slice.split(":")
        path = BRICK_PATH % device
        _exec_command("mkdir %s" % path, host)
        bricks.append("%s:%s" % (host, path))
    print(bricks)
    return bricks


def _volume_create(gluster_host, volume_name, bricks):
    create_base = "gluster volume create %s %s"
    command = create_base % (volume_name, " ".join(bricks))
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


def setup_volume(gluster_host, volume_name, slices):
    print("start: setup volume")
    bricks = _create_bricks(slices)
    _volume_create(gluster_host, volume_name, bricks)
    _volume_start(gluster_host, volume_name)


def volume_remove(gluster_host, volume_name):
    # TODO(johngarbutt) might be cleaner to clean all bricks here?
    print("start: volume remove")
    _volume_stop(gluster_host, volume_name)
    _volume_delete(gluster_host, volume_name)


def clean_brick(gluster_host, device_name):
    print("start: clean brick")
    path = BRICK_PATH % device_name
    _exec_command("rm -rf %s" % path, gluster_host)
