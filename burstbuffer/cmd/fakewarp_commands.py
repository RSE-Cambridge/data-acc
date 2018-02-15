
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

import json
import sys

from cliff.command import Command


class Pools(Command):
    """Output burst buffer pools"""

    def take_action(self, parsed_args):
        fake_pools = {
            "pools": [
                {"id": "dwcache", "units": "bytes", "granularity": 16777216,
                 "quantity": 2048, "free": 2048},
                {"id": "test_pool", "units": "bytes", "granularity": 16777216,
                 "quantity": 2048, "free": 2048}
            ]
        }
        json.dump(fake_pools, sys.stdout, sort_keys=True, indent=4)
