#!/bin/bash

set -eux

make

scp ./bin/amd64/data-acc-brick-host centos@128.232.224.186:~/
scp ./data-acc-brick-host.service centos@128.232.224.186:~/

cat >install.sh <<EOF
#!/bin/bash
sudo chmod 755 data-acc-brick-host.service
sudo cp data-acc-brick-host.service /lib/systemd/system/.
sudo sudo systemctl enable data-acc-brick-host.service
sudo systemctl start data-acc-brick-host
sudo journalctl -f -u data-acc-brick-host
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
