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

import io
import mock
import tempfile

import fixtures
import testtools

from burstbuffer.cmd import fakewarp


class TestFakeWarp(testtools.TestCase):

    def setUp(self):
        super(TestFakeWarp, self).setUp()
        stdout = self.useFixture(fixtures.StringStream('stdout')).stream
        self.useFixture(fixtures.MonkeyPatch('sys.stdout', stdout))
        stderr = self.useFixture(fixtures.StringStream('stderr')).stream
        self.useFixture(fixtures.MonkeyPatch('sys.stderr', stderr))

    def test_pools(self):
        result = fakewarp.main(["--function", "pools"])
        self.assertEqual(0, result)

    def test_show_instances(self):
        result = fakewarp.main(["--function", "show_instances"])
        self.assertEqual(0, result)

    def test_show_sessions(self):
        result = fakewarp.main(["--function", "show_sessions"])
        self.assertEqual(0, result)

    def test_teardown(self):
        cmdline = "--function teardown --token 347 --job /tmp/fakescript"
        result = fakewarp.main(cmdline.split(" "))
        self.assertEqual(0, result)

    def test_teardown_with_hurry(self):
        cmdline = "--function teardown --token 347 --job /tmp/fakescript"
        cmdline += " --hurry"
        result = fakewarp.main(cmdline.split(" "))
        self.assertEqual(0, result)

    def test_job_process(self):
        stdout = io.StringIO()
        self.useFixture(fixtures.MonkeyPatch('sys.stdout', stdout))
        with tempfile.NamedTemporaryFile() as f:
            f.write(b"#!/bin/bash\n")
            f.write(b"#DW jobdw capacity=1GiB")
            f.flush()
            cmdline = "--function job_process --job %s" % f.name

            result = fakewarp.main(cmdline.split(" "))

            self.assertEqual(0, result)
            self.assertEqual("capacity=1GiB\n", stdout.getvalue())

    def test_setup(self):
        cmdline = "--function setup "
        cmdline += "--token 13 --caller SLURM --user 995 "
        cmdline += "--groupid 995 --capacity dwcache:1GiB "
        cmdline += "--job /var/lib/slurmd/hash.3/job.13/script"
        result = fakewarp.main(cmdline.split(" "))
        self.assertEqual(0, result)

    def test_real_size(self):
        cmdline = "--function real_size --token 13"
        result = fakewarp.main(cmdline.split(" "))
        self.assertEqual(0, result)

    @mock.patch("time.sleep")
    def test_data_in(self, mock_sleep):
        cmdline = "--function data_in --token 15 --job /tmp/jobscript"
        result = fakewarp.main(cmdline.split(" "))
        self.assertEqual(0, result)

    def test_paths(self):
        with tempfile.NamedTemporaryFile() as f:
            cmdline = "--function paths "
            cmdline += "--job /tmp/18/scirpt --token 18 --pathfile %s" % f.name
            result = fakewarp.main(cmdline.split(" "))
            self.assertEqual(0, result)
            self.assertEqual(b'DW_PATH_TEST=/tmp/dw', f.readlines()[0])

    def test_pre_run(self):
        with tempfile.NamedTemporaryFile() as f:
            f.write(b"somecomputehostname")
            f.flush()
            cmdline = "pre_run --token 19 --job /tmp/script "
            cmdline += "--nodehostnamefile %s" % f.name

            result = fakewarp.main(cmdline.split(" "))

            self.assertEqual(0, result)

    def test_post_run(self):
        cmdline = "post_run --token 27 --job /tmp/script"
        result = fakewarp.main(cmdline.split(" "))
        self.assertEqual(0, result)

    def test_data_out(self):
        cmdline = "--function data_out --token 28 --job /tmp/script"
        result = fakewarp.main(cmdline.split(" "))
        self.assertEqual(0, result)

    def test_show_configurations(self):
        cmdline = "--function show_configurations"
        result = fakewarp.main(cmdline.split(" "))
        self.assertEqual(0, result)
