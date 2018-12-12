set -eux

set -a
. /etc/data-acc/dacd.conf
set +a

/usr/local/bin/etcdctl --key /etc/data-acc/pki/`hostname`.dac.hpc.cam.ac.uk-key.pem --cert /etc/data-acc/pki/`hostname`.dac.hpc.cam.ac.uk.pem --cacert /etc/data-acc/pki/ca.pem del --prefix ''

# Kill all lustre filesystems
ssh dac1 sudo umount -at lustre
ssh dac2 sudo umount -at lustre
ssh dac3 sudo umount -at lustre

ssh dac1 sudo systemctl restart dacd
ssh dac2 sudo systemctl restart dacd
ssh dac3 sudo systemctl restart dacd

/usr/local/bin/etcdctl --key /etc/data-acc/pki/`hostname`.dac.hpc.cam.ac.uk-key.pem --cert /etc/data-acc/pki/`hostname`.dac.hpc.cam.ac.uk.pem --cacert /etc/data-acc/pki/ca.pem get --prefix ''
