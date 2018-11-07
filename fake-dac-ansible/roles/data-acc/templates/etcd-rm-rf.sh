set -eux

set -a
. /etc/data-acc/dacd.conf
set +a

/usr/local/bin/etcdctl --key /etc/data-acc/pki/`hostname`.dac.hpc.cam.ac.uk-key.pem --cert /etc/data-acc/pki/`hostname`.dac.hpc.cam.ac.uk.pem --cacert /etc/data-acc/pki/ca.pem del --prefix ''
