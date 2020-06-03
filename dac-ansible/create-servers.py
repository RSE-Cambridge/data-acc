#!/usr/bin/env python

import openstack
from optparse import OptionParser

parser = OptionParser()
parser.add_option("-k", "--key", dest="key",
                  help="SSH key pair name", metavar="KEYPAIR_NAME")
parser.add_option("-n", "--network", dest="network",
                  help="Network name", metavar="NETWORK_NAME")
(options, args) = parser.parse_args()

IMAGE_NAME = "CentOS-7-x86_64-GenericCloud"
FLAVOR_NAME = "C1.vss.tiny"
NETWORK_NAME = options.network
KEYPAIR_NAME = options.key

def get_connection():
    # openstack.enable_logging(debug=True)
    conn = openstack.connect()
    return conn


def create_server(conn, name, image, flavor, network):
    server = conn.compute.find_server(name)
    if server is None:
        server = conn.compute.create_server(
            name=name, image_id=image.id, flavor_id=flavor.id,
            networks=[{"uuid": network.id}], key_name=KEYPAIR_NAME)
        server = conn.compute.wait_for_server(server)

    details = conn.compute.get_server(server.id)
    return details.addresses[NETWORK_NAME][0]['addr']


def main():
    conn = get_connection()

    image = conn.compute.find_image(IMAGE_NAME)
    if image is None:
        raise Exception("Can't find %s" % IMAGE_NAME)
    flavor = conn.compute.find_flavor(FLAVOR_NAME)
    if flavor is None:
        raise Exception("Can't find %s" % FLAVOR_NAME)
    network = conn.network.find_network(NETWORK_NAME)
    if network is None:
        raise Exception("Can't find %s" % NETWORK_NAME)

    servers = {}
    servers['dac1'] = create_server(conn, 'dac1', image, flavor, network)
    servers['dac2'] = create_server(conn, 'dac2', image, flavor, network)
    servers['dac3'] = create_server(conn, 'dac3', image, flavor, network)
    servers['dac-etcd'] = create_server(
            conn, 'dac-etcd', image, flavor, network)
    servers['dac-slurm-master'] = create_server(
            conn, 'dac-slurm-master', image, flavor, network)
    servers['slurm-cpu1'] = create_server(
            conn, 'slurm-cpu1', image, flavor, network)
    servers['slurm-cpu2'] = create_server(
            conn, 'slurm-cpu2', image, flavor, network)

    inventory_template = """[dac_workers]
dac1.dac.hpc.cam.ac.uk ansible_host=%s ansible_user=centos
dac2.dac.hpc.cam.ac.uk ansible_host=%s ansible_user=centos
dac3.dac.hpc.cam.ac.uk ansible_host=%s ansible_user=centos

[etcd_master]
dac-etcd.dac.hpc.cam.ac.uk ansible_host=%s ansible_user=centos

[etcd:children]
etcd_master
dac_workers

[nfs:children]
slurm_master

[openstack:children]
etcd
slurm

[slurm_master]
dac-slurm-master.dac.hpc.cam.ac.uk ansible_host=%s ansible_user=centos

[slurm_workers]
slurm-cpu1.dac.hpc.cam.ac.uk ansible_host=%s ansible_user=centos
slurm-cpu2.dac.hpc.cam.ac.uk ansible_host=%s ansible_user=centos

[slurm:children]
slurm_master
slurm_workers"""

    print(inventory_template % (
            servers['dac1'],
            servers['dac2'],
            servers['dac3'],
            servers['dac-etcd'],
            servers['dac-slurm-master'],
            servers['slurm-cpu1'],
            servers['slurm-cpu2']))


if __name__ == '__main__':
    main()
