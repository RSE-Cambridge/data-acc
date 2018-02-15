
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


def _output_as_json(output):
    json.dump(output, sys.stdout, sort_keys=True, indent=4)


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
        _output_as_json(fake_pools)


class ShowInstances(Command):
    """Show burst buffers instances"""

    def take_action(self, parsed_args):
        fake_instances = [
            {"capacity": {
                "bytes": 1099511627776, "nodes": 2},
             "created": 1478231657, "expiration": 0, "expired": False,
             "id": 74, "intact": True, "label": "alpha",
             "limits": {
                 "write_window_length": 86400, "write_window_multiplier": 10},
             "links": {
                 "configurations": [72], "session": 74},
             "public": True,
             "state": {
                 "actualized": True, "fuse_blown": False, "goal": "create",
                 "mixed": False, "transitioning": False}},
            {"capacity": {
                "bytes": 1099511627776, "nodes": 2},
             "created": 1478232104, "expiration": 0, "expired": False,
             "id": 75, "intact": True, "label": "I75-0",
             "limits": {
                "write_window_length": 86400, "write_window_multiplier": 10},
             "links": {
                "configurations": [73], "session": 75},
             "public": False,
             "state": {
                "actualized": True, "fuse_blown": False, "goal": "create",
                "mixed": False, "transitioning": False}},
        ]
        fake_instances = {"instances": fake_instances}
        _output_as_json(fake_instances)


class ShowSessions(Command):
    """Show burst buffers sessions"""

    def take_action(self, parsed_args):
        fake_sessions = [
            {"created": 1478231657, "creator": "CLI", "expiration": 0,
             "expired": False, "id": 74,
             "links": {"client_nodes": []},
             "owner": 1001,
             "state": {
                 "actualized": True, "fuse_blown": False, "goal": "create",
                 "mixed": False, "transitioning": False},
             "token": "alpha"},
            {"created": 1478232104, "creator": "SLURM", "expiration": 0,
             "expired": False, "id": 75,
             "links": {"client_nodes": ["nid00039"]},
             "owner": 1001,
             "state": {
                 "actualized": True, "fuse_blown": False, "goal": "create",
                 "mixed": False, "transitioning": False},
             "token": "347"},
        ]
        fake_sessions = {"sessions": fake_sessions}
        _output_as_json(fake_sessions)
