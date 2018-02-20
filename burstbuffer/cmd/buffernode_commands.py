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
import socket

from cliff.command import Command


class Startup(Command):
    """On node start tidy up any orphans, etc."""

    def take_action(self, parsed_args):
        hostname = socket.gethostname()
        self.app.LOG.info("start of day for %s" % hostname)


class Event(Command):
    """Callback for when a hostslice event occurs"""

    def take_action(self, parsed_args):
        self.app.LOG.info("event occured")
        for key in os.environ:
            if key.startswith("ETCD_WATCH_"):
                short_key = key.trim("ETCD_WATCH_")
                print("%s: %s" % (short_key, os.environ[key]))
