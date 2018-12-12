set -eux

# NOTE: must be run as the dac user

set -a
. /etc/data-acc/dacd.conf
set +a

# Ensure dacctl can write to its log
sudo touch /var/log/dacctl.log
sudo chown dac /var/log/dacctl.log

/usr/local/bin/dacctl show_sessions
/usr/local/bin/dacctl show_instances
/usr/local/bin/dacctl pools

sleep 5

token="test${RANDOM}"
/usr/local/bin/dacctl create_persistent --token $token -c johng -C default:30GB --user 1001 -g 1001 -a striped -T scratch

sleep 5

/usr/local/bin/etcdctl --key /etc/data-acc/pki/`hostname`.dac.hpc.cam.ac.uk-key.pem --cert /etc/data-acc/pki/`hostname`.dac.hpc.cam.ac.uk.pem --cacert /etc/data-acc/pki/ca.pem get --prefix ''

sleep 15

/usr/local/bin/dacctl show_sessions
/usr/local/bin/dacctl show_instances
/usr/local/bin/dacctl pools

sleep 15

/usr/local/bin/dacctl teardown --job /usr/local/bin/data-acc-test.sh --token $token

sleep 15

/usr/local/bin/etcdctl --key /etc/data-acc/pki/`hostname`.dac.hpc.cam.ac.uk-key.pem --cert /etc/data-acc/pki/`hostname`.dac.hpc.cam.ac.uk.pem --cacert /etc/data-acc/pki/ca.pem get --prefix ''
