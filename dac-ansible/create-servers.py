#!/usr/bin/env python

import argparse

from openstack import exceptions
import openstack


IMAGE_NAME = "CentOS-7-x86_64-GenericCloud"
FLAVOR_NAME = "C1.vss.tiny"
NETWORK_NAME = "WCDC-Data43"
KEYPAIR_NAME = "usual"


def get_connection():
    #openstack.enable_logging(debug=True)
    conn = openstack.connect()
    return conn


def create_server(conn, name):
    try:
        return conn.compute.get_server(name)
    except exceptions.ResourceNotFound:
        pass

    server = conn.compute.create_server(
        name=SERVER_NAME, image_id=image.id, flavor_id=flavor.id,
        networks=[{"uuid": network.id}], key_name=keypair.name)
    return server


def main():
    conn = get_connection()

    image = conn.compute.find_image(IMAGE_NAME)
    flavor = conn.compute.find_flavor(FLAVOR_NAME)
    network = conn.network.find_network(NETWORK_NAME)

    servers = {}
    servers['dac1', create_server(conn, 'dac1')]
    servers['dac2', create_server(conn, 'dac2')]

    print "[dac-workers]"
    print "dac1.dac.hpc.cam.ac.uk ansible_host=%s ansible_user=centos" % servers['dac1'].access_ipv4
    print "dac2.dac.hpc.cam.ac.uk ansible_host=%s ansible_user=centos" % servers['dac2'].access_ipv4


if __name__ == '__main__':
    main()
