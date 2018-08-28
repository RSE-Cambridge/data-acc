#!/usr/bin/env python2
import yaml
from collections import OrderedDict

def main():
    
    dac_no = 2
    #init yaml structure
    dac = {'dac-prod':{'children': {'fs1001':{'hosts':{}}}}}

    #create nvme,index as orderd dict to preserve lustre global index order in relation to nvme
    for i in range(1,dac_no+1):
        dac['dac-prod']['children']['fs1001']['hosts']["dac-e-{}".format(i)] = {'fs1001_osts': OrderedDict()}

    #set dac 1 as mdt and mgs
    dac['dac-prod']['children']['fs1001']['hosts']['dac-e-1']['fs1001_mgs'] = "nvme0n1"
    dac['dac-prod']['children']['fs1001']['hosts']['dac-e-1']['fs1001_mdt'] = "nvme0n1"

    #create keys for nvmes and index init to 0
    for i in range(1,12):
        dac['dac-prod']['children']['fs1001']['hosts']['dac-e-1']['fs1001_osts']["nvme{}n1".format(i)] = 0
 
    for i in range(2,dac_no+1):
        for j in range(0,12):
           dac['dac-prod']['children']['fs1001']['hosts']["dac-e-{}".format(i)]['fs1001_osts']["nvme{}n1".format(j)] = 0

    #globaly index all nvmes
    n = 0
    for i in range(1,dac_no+1):
        for key in dac['dac-prod']['children']['fs1001']['hosts']["dac-e-{}".format(i)]['fs1001_osts']:
            dac['dac-prod']['children']['fs1001']['hosts']["dac-e-{}".format(i)]['fs1001_osts'][key] = n
            n+=1

    #cast orderdict back to dict for yaml
    for i in range(1,dac_no+1):
        dac['dac-prod']['children']['fs1001']['hosts']["dac-e-{}".format(i)]['fs1001_osts'] = dict(dac['dac-prod']['children']['fs1001']['hosts']["dac-e-{}".format(i)]['fs1001_osts'])

    dac['dac-prod']['children']['fs1001']['vars'] = {
        'fs1001_mgsnode': 'dac-e-1',
        'fs1001_client_port': '10001'
    }

    print(yaml.dump(dac))


if __name__ == "__main__":
    main()
