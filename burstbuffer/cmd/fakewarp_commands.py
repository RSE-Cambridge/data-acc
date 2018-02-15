
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
import time

from cliff.command import Command


def _output_as_json(cmd, output):
    json.dump(output, cmd.app.stdout, sort_keys=True, indent=4)


class Pools(Command):
    """Output burst buffer pools"""

    def take_action(self, parsed_args):
        fake_pools = {
            "pools": [
                {"id": "dwcache", "units": "bytes",
                 "granularity": 1073741824,  # 1GB
                 "quantity": 4096,
                 "free": 4096},  # i.e. 4TB
                {"id": "test_pool", "units": "bytes", "granularity": 16777216,
                 "quantity": 2048, "free": 2048}
            ]
        }
        _output_as_json(self, fake_pools)


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
        _output_as_json(self, fake_instances)


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
        _output_as_json(self, fake_sessions)


class Teardown(Command):
    """Start the teardown of the given burst buffer"""

    def get_parser(self, prog_name):
        parser = super(Teardown, self).get_parser(prog_name)
        parser.add_argument('--token', type=str, dest="job_id",
                            help="Job ID")
        parser.add_argument('--job', type=str, dest="buffer_script",
                            help="Path to burst buffer script file.")
        parser.add_argument('--hurry', action="store_true", default=False)
        return parser

    def take_action(self, parsed_args):
        print(parsed_args.job_id)
        print(parsed_args.buffer_script)
        print(parsed_args.hurry)


class JobProcess(Command):
    """Initial call when job is run to parse buffer script."""

    def get_parser(self, prog_name):
        parser = super(JobProcess, self).get_parser(prog_name)
        parser.add_argument('--job', type=str, dest="buffer_script",
                            help="Path to burst buffer script file.")
        return parser

    def take_action(self, parsed_args):
        job_config_line = None
        with open(parsed_args.buffer_script) as f:
            for line in f:
                if line.startswith("#DW jobdw"):
                    job_config_line = line
                    break

        config = job_config_line.strip("#DW jobdw ")
        print(config)
        # this validates the buffer script, next step is calling "setup"
        # once there is enough available bust buffer space


class Setup(Command):
    """Create the burst buffer, ready to start the data stage_in"""

    def get_parser(self, prog_name):
        parser = super(Setup, self).get_parser(prog_name)
        parser.add_argument('--token', type=str, dest="job_id",
                            help="Job ID")
        parser.add_argument('--job', type=str, dest="buffer_script",
                            help="Path to burst buffer script file.")
        parser.add_argument('--caller', type=str,
                            help="Caller, i.e. SLURM")
        parser.add_argument('--user', type=int,
                            help="User id, i.e. 1001")
        parser.add_argument('--groupid', type=int,
                            help="Group id, i.e. 1001")
        parser.add_argument('--capacity', type=str,
                            help="The pool and capacity, i.e. dwcache:1GiB")
        return parser

    def take_action(self, parsed_args):
        # this should add the burst buffer in the DB, so real_size works
        print(parsed_args.job_id)
        print(parsed_args.buffer_script)
        print(parsed_args.capacity)
        print("pool: %s, capacity: %s" % tuple(
            parsed_args.capacity.split(":")))


class RealSize(Command):
    """Report actual size of burst buffer, rounded up for granularity"""

    def get_parser(self, prog_name):
        parser = super(RealSize, self).get_parser(prog_name)
        parser.add_argument('--token', type=str, dest="job_id",
                            help="Job ID")
        return parser

    def take_action(self, parsed_args):
        fake_size = {
            "token": parsed_args.job_id,
            "capacity": 17592186044416,
            "units": "bytes"
        }
        _output_as_json(self, fake_size)


class DataIn(Command):
    """Start copy of data into the burst buffer"""

    def get_parser(self, prog_name):
        parser = super(DataIn, self).get_parser(prog_name)
        parser.add_argument('--token', type=str, dest="job_id",
                            help="Job ID")
        parser.add_argument('--job', type=str, dest="buffer_script",
                            help="Path to burst buffer script file.")
        return parser

    def take_action(self, parsed_args):
        # although I think this is async...
        time.sleep(10)
        # "No matching session" if there is no matching job found


class Paths(Command):
    """Start copy of data into the burst buffer"""

    def get_parser(self, prog_name):
        parser = super(Paths, self).get_parser(prog_name)
        parser.add_argument('--token', type=str, dest="job_id",
                            help="Job ID")
        parser.add_argument('--job', type=str, dest="buffer_script",
                            help="Path to burst buffer script file.")
        parser.add_argument('--pathfile', type=str,
                            help="Path to write out environment variables.")
        return parser

    def take_action(self, parsed_args):
        with open(parsed_args.pathfile, "w") as f:
            f.write("DW_PATH_TEST=/tmp/dw")
