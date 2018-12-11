#!/usr/bin/env python

import openstack


IMAGE_NAME = "CentOS-7-x86_64-GenericCloud"
FLAVOR_NAME = "C1.vss.tiny"
NETWORK_NAME = "WCDC-Data43"
KEYPAIR_NAME = "usual"


def get_connection():
    # openstack.enable_logging(debug=True)
    conn = openstack.connect()
    return conn


def create_server(conn, name, image, flavor, network):
    server = conn.compute.find_server(name)
    if server is not None:
        return server

    server = conn.compute.create_server(
        name=name, image_id=image.id, flavor_id=flavor.id,
        networks=[{"uuid": network.id}], key_name=KEYPAIR_NAME)
    return server


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

    inventory_template = """[dac-workers]
dac1.dac.hpc.cam.ac.uk ansible_host=%s ansible_user=centos
dac2.dac.hpc.cam.ac.uk ansible_host=%s ansible_user=centos
dac3.dac.hpc.cam.ac.uk ansible_host=%s ansible_user=centos

[etcd-master]
dac-etcd.dac.hpc.cam.ac.uk ansible_host=%s ansible_user=centos

[etcd:children]
etcd-master
dac-workers

[slurm-master]
dac-slurm-master.dac.hpc.cam.ac.uk ansible_host=%s ansible_user=centos

[slurm-workers]
slurm-cpu1.dac.hpc.cam.ac.uk ansible_host=%s ansible_user=centos
slurm-cpu2.dac.hpc.cam.ac.uk ansible_host=%s ansible_user=centos

[slurm:children]
slurm-master
slurm-workers
"""
    print inventory_template % (
            servers['dac1'].access_ipv4,
            servers['dac2'].access_ipv4,
            servers['dac3'].access_ipv4,
            servers['dac-etcd'].access_ipv4,
            servers['dac-slurm-master'].access_ipv4,
            servers['slurm-cpu1'].access_ipv4,
            servers['slurm-cpu2'].access_ipv4)


if __name__ == '__main__':
    main()
