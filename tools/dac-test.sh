set -eux

set -a
. /etc/data-acc/dacd.conf
set +a

/usr/local/bin/fakewarp show_sessions
/usr/local/bin/fakewarp show_instances
/usr/local/bin/fakewarp pools

sleep 5

token="test${RANDOM}"
/usr/local/bin/fakewarp create_persistent --token $token -c johng -C default:30GB --user 1001 -g 1001 -a striped -T scratch

sleep 5

/usr/local/bin/etcdctl --key /etc/data-acc/pki/`hostname`.dac.hpc.cam.ac.uk-key.pem --cert /etc/data-acc/pki/`hostname`.dac.hpc.cam.ac.uk.pem --cacert /etc/data-acc/pki/ca.pem get --prefix ''

sleep 15

/usr/local/bin/fakewarp show_sessions
/usr/local/bin/fakewarp show_instances
/usr/local/bin/fakewarp pools

sleep 15

/usr/local/bin/fakewarp teardown --job /usr/local/bin/data-acc-test.sh --token $token

sleep 15

/usr/local/bin/etcdctl --key /etc/data-acc/pki/`hostname`.dac.hpc.cam.ac.uk-key.pem --cert /etc/data-acc/pki/`hostname`.dac.hpc.cam.ac.uk.pem --cacert /etc/data-acc/pki/ca.pem get --prefix ''
