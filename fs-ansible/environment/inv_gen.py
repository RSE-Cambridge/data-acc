#!/bin/env python
import yaml
from collections import OrderedDict

def main():
    
    dac_no = 1
    #init yaml structure
    dac = {'all':{'children': {'fs1001':{'hosts':{}}}}}

    #create nvme,index as orderd dict to preserve lustre global index order in relation to nvme
    for i in range(1,dac_no+1):
        dac['all']['children']['fs1001']['hosts']["dac-e-{}".format(i)] = {'osts': OrderedDict()}

    #set dac 1 as mdt
    dac['all']['children']['fs1001']['hosts']['dac-e-1']['mdts'] = ["nvme0n1"]

    #create keys for nvmes and index init to 0
    for i in range(1,12):
        dac['all']['children']['fs1001']['hosts']['dac-e-1']['osts']["nvme{}n1".format(i)] = 0
 
    for i in range(2,dac_no+1):
        for j in range(0,12):
           dac['all']['children']['fs1001']['hosts']["dac-e-{}".format(i)]['osts']["nvme{}n1".format(j)] = 0

    #globaly index all nvmes
    n = 0
    for i in range(1,dac_no+1):
        for key in dac['all']['children']['fs1001']['hosts']["dac-e-{}".format(i)]['osts']:
            dac['all']['children']['fs1001']['hosts']["dac-e-{}".format(i)]['osts'][key] = n
            n+=1

    #cast orderdict back to dict for yaml
    for i in range(1,dac_no+1):
        dac['all']['children']['fs1001']['hosts']["dac-e-{}".format(i)]['osts'] = dict(dac['all']['children']['fs1001']['hosts']["dac-e-{}".format(i)]['osts'])

    dac['all']['children']['fs1001']['vars'] = {'mgsnode': '10.47.18.1@o2ib1'}

    print(yaml.dump(dac))



if __name__ == "__main__":
    main()
