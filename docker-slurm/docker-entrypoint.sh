#!/bin/bash
set -e

MYSQL_HOST=${MYSQL_HOST:-"192.168.0.109"}
SLURM_DB=${SLURM_DB:-"192.168.0.109"}
SLURM_CTL=${SLURM_CTL:-"192.168.0.109"}

# update config
cat /etc/slurm/slurm.conf.tmpl | envsubst > /etc/slurm/slurm.conf
cat /etc/slurm/slurmdbd.conf.tmpl | envsubst > /etc/slurm/slurmdbd.conf

if [ "$1" = "slurmdbd" ]
then
    echo "---> Starting the MUNGE Authentication service (munged) ..."
    gosu munge /usr/sbin/munged

    echo "---> Starting the Slurm Database Daemon (slurmdbd) ..."

    until echo "SELECT 1" | mysql -h $MYSQL_HOST -uslurm -ppassword 2>&1 > /dev/null
    do
        echo "-- Waiting for database to become active ..."
        sleep 2
    done
    echo "-- Database is now active ..."

    exec gosu slurm /usr/sbin/slurmdbd -Dvvv
fi

if [ "$1" = "slurmctld" ]
then
    echo "---> Starting the MUNGE Authentication service (munged) ..."
    gosu munge /usr/sbin/munged

    echo "---> Waiting for slurmdbd to become active before starting slurmctld ..."

    until 2>/dev/null >/dev/tcp/${SLURM_DB}/6819
    do
        echo "-- slurmdbd is not available.  Sleeping ..."
        sleep 2
    done
    echo "-- slurmdbd is now active ..."

    echo "---> Starting the Slurm Controller Daemon (slurmctld) ..."
    exec gosu slurm /usr/sbin/slurmctld -Dvvv -i
fi

if [ "$1" = "slurmd" ]
then
    echo "---> Starting the MUNGE Authentication service (munged) ..."
    gosu munge /usr/sbin/munged

    echo "---> Waiting for slurmctld to become active before starting slurmd..."

    until 2>/dev/null >/dev/tcp/${SLURM_CTL}/6817
    do
        echo "-- slurmctld is not available.  Sleeping ..."
        sleep 2
    done
    echo "-- slurmctld is now active ..."

    echo "---> Starting the Slurm Node Daemon (slurmd) ..."
    exec /usr/sbin/slurmd -Dvvv
fi

exec "$@"
