#!/bin/env python
import yaml
import sys
import argparse
from collections import OrderedDict

def build_inv(dac_no,nvme_no,fs,fstype):
    """Build shape of the DAC for Lustre.
        takes:
            dac_no = number of requierd nodes
            nvme_no = number of nvmes indexed from 0
                eg. 12 nvmes is nvme_no = 11
    """
    #init yaml structure
    dac = {'all':{'children': {fs:{'hosts':{}}}}}

    #create nvme,index as orderd dict to preserve lustre global index order in relation to nvme
    for i in range(1,dac_no+1):
        dac['all']['children'][fs]['hosts']["dac-e-{}".format(i)] = {fs+"_osts": OrderedDict()}

    #set dac 1 as mdt
    if fstype == 'beegfs':
        dac['all']['children'][fs]['hosts']['dac-e-1']['mds'] = "nvme0n1"
    else:
        dac['all']['children'][fs]['hosts']['dac-e-1']["{}_mdt".format(fs)] = {"nvme0n1": 0}
        dac['all']['children'][fs]['hosts']['dac-e-1']["{}_mgs".format(fs)] = "sdb"

    #broken but passable way to ensure the nvmes are alocated evenly acrose numa domain
    half = int(nvme_no/2)
    n = range(0,12)
    a = 0
    nvme=[]
    for i in [1,2]:
        nvme.append(n[a+i:a+half+i])
        a+=half
    nvme = [n for x in nvme for n in x]

    #create keys for nvmes and index init to 0
    for i in nvme:
        dac['all']['children'][fs]['hosts']['dac-e-1']["{}_osts".format(fs)]["nvme{}n1".format(i)] = 0
 
    for i in range(2,dac_no+1):
        for j in nvme:
            dac['all']['children'][fs]['hosts']["dac-e-{}".format(i)]["{}_osts".format(fs)]["nvme{}n1".format(j)] = 0
        #uncoment bellow if you want to enable/disable dne
        if i > 0 and i <= 24:
            dac['all']['children'][fs]['hosts']["dac-e-{}".format(i)]["{}_mdt".format(fs)] = {"nvme0n1": (i-2)}

    #globaly index all nvmes
    n = 0
    for i in range(1,dac_no+1):
        for key in dac['all']['children'][fs]['hosts']["dac-e-{}".format(i)]["{}_osts".format(fs)]:
            dac['all']['children'][fs]['hosts']["dac-e-{}".format(i)]["{}_osts".format(fs)][key] = n
            n+=1

    #cast orderdict back to dict for yaml
    for i in range(1,dac_no+1):
        dac['all']['children'][fs]['hosts']["dac-e-{}".format(i)]["{}_osts".format(fs)] = dict(dac['all']['children'][fs]['hosts']["dac-e-{}".format(i)]["{}_osts".format(fs)])

    if fstype == 'beegfs':
        dac['all']['children'][fs]['hosts']['dac-e-1']['mgs'] = 'true'
        dac['all']['children'][fs]['vars'] = {'mgsnode': '10.47.18.1'}
    else:
        dac['all']['children'][fs]['vars'] = {"{}_mgsnode".format(fs): '10.47.18.1@o2ib1'}


    # sort out the ip of the hfis. this is requierd currently for seting up lustre multirail routs
    # and arp
    dac_ip = {}
    j = 25
    for i in range(1,dac_no+1):
        dac_ip['dac-e-'+str(i)] = {'ib0': "10.47.18."+str(i), 'ib1': "10.47.18."+str(j)}
        dac['all']['children'][fs]['hosts']["dac-e-{}".format(i)]['ip'] = dac_ip["dac-e-{}".format(i)]
        j+=1

    print(yaml.dump(dac))


if __name__ == "__main__":

    parser = argparse.ArgumentParser(
        description='Test program for building shape of DAC storage for number of nodes, and nvmes')
    parser.add_argument('--dac', type=int, help='Set number of DAC nodes')
    parser.add_argument('--nvme', type=int, help='Set number of NVMe index from 0')
    parser.add_argument('--fsname', type=str, help='Set fs name', default='fs1001')
    parser.add_argument('--fstype', type=str, help='Set buffer fs type')
    if len(sys.argv) == 1:
        parser.print_help(sys.stderr)
        sys.exit(1)
    args = parser.parse_args()
    build_inv(args.dac, args.nvme, args.fsname, args.fstype)

