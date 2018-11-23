#!/bin/bash

set -eux

make

scp ./bin/amd64/dacd centos@128.232.224.186:~/
scp ./dacd.service centos@128.232.224.186:~/

cat >install.sh <<EOF
#!/bin/bash
set -eux
sudo chmod 755 dacd.service
sudo cp dacd.service /lib/systemd/system/.
sudo sudo systemctl enable dacd.service
sudo systemctl start dacd

sudo systemctl stop dacd
scp slurm-master:~/dacd .
sudo systemctl start dacd

sudo journalctl -f -u dacd
EOF
chmod +x install.sh
scp ./install.sh centos@128.232.224.186:~/
rm -f install.sh

cd docker-slurm

rm -rf ./bin
mkdir -p ./bin
cp -r ../bin/amd64/* ./bin

docker-compose build
docker-compose push
