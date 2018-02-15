#!/usr/bin/env python

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


from setuptools import setup, find_packages


PROJECT = 'burstbuffer'
VERSION = '0.1.0'

try:
    long_description = open('README.md', 'rt').read()
except IOError:
    long_description = ''

setup(
    name=PROJECT,
    version=VERSION,

    description='Burst Buffer using commodity hardware',
    long_description=long_description,

    author='StackHPC',
    author_email='john.garbutt@stackhpc.com',

    url='https://github.com/johngarbutt/burstbuffer',
    download_url='https://github.com/johngarbutt/burstbuffer/tarball/master',

    provides=[],
    install_requires=open('requirements.txt', 'rt').read().splitlines(),

    namespace_packages=[],
    packages=find_packages(),
    include_package_data=True,

    entry_points={
        'console_scripts': [
            'fakewarp = burstbuffer.cmd.fakewarp:main',
        ],
        'burstbuffer.fakewarp': [
            'pools = burstbuffer.cmd.fakewarp_commands:Pools',
            'show_instances = burstbuffer.cmd.fakewarp_commands:ShowInstances',
            'show_sessions = burstbuffer.cmd.fakewarp_commands:ShowSessions',
            'teardown = burstbuffer.cmd.fakewarp_commands:Teardown',
            'job_process = burstbuffer.cmd.fakewarp_commands:JobProcess',
            'setup = burstbuffer.cmd.fakewarp_commands:Setup',
            'real_size = burstbuffer.cmd.fakewarp_commands:RealSize',
        ],
    },

    zip_safe=False,
)
