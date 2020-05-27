FROM centos:7

# Ideas mostly from:
# https://github.com/giovtorres/slurm-docker-cluster/releases/tag/18.08.6

LABEL org.label-schema.docker.cmd="docker-compose up -d" \
      org.label-schema.name="slurm-docker-cluster" \
      org.label-schema.description="Slurm Docker cluster on CentOS 7" \
      maintainer="John Garbutt"

ARG SLURM_TAG=slurm-19-05-7-1
ARG GOSU_VERSION=1.12

RUN set -ex \
    && yum makecache fast \
    && yum -y update \
    && yum -y install epel-release \
    && yum -y install \
       wget \
       bzip2 \
       perl \
       gcc \
       gcc-c++\
       git \
       json-c \
       json-c-devel \
       gnupg \
       make \
       munge \
       munge-devel \
       python-devel \
       python-pip \
       python-virtualenv \
       mariadb-server \
       mariadb-devel \
       psmisc \
       bash-completion \
       vim-enhanced \
    && yum clean all \
    && rm -rf /var/cache/yum

RUN set -ex \
    && wget -O /usr/local/bin/gosu "https://github.com/tianon/gosu/releases/download/$GOSU_VERSION/gosu-amd64" \
    && wget -O /usr/local/bin/gosu.asc "https://github.com/tianon/gosu/releases/download/$GOSU_VERSION/gosu-amd64.asc" \
    && export GNUPGHOME="$(mktemp -d)" \
    && gpg --keyserver ha.pool.sks-keyservers.net --recv-keys B42F6819007F00F88E364FD4036A9C25BF357DD4 \
    && gpg --batch --verify /usr/local/bin/gosu.asc /usr/local/bin/gosu \
    && rm -rf "${GNUPGHOME}" /usr/local/bin/gosu.asc \
    && chmod +x /usr/local/bin/gosu \
    && gosu nobody true

RUN set -x \
    && git clone https://github.com/SchedMD/slurm.git \
    && pushd slurm \
    && git checkout tags/$SLURM_TAG \
    && ./configure --enable-debug --prefix=/usr --sysconfdir=/etc/slurm \
        --with-mysql_config=/usr/bin  --libdir=/usr/lib64 \
    && make install \
    && install -D -m644 etc/cgroup.conf.example /etc/slurm/cgroup.conf.example \
    && install -D -m644 etc/slurm.conf.example /etc/slurm/slurm.conf.example \
    && install -D -m644 etc/slurmdbd.conf.example /etc/slurm/slurmdbd.conf.example \
    && install -D -m644 contribs/slurm_completion_help/slurm_completion.sh /etc/profile.d/slurm_completion.sh \
    && popd \
    && rm -rf slurm \
    && groupadd -r --gid=995 slurm \
    && useradd -r -g slurm --uid=995 slurm \
    && mkdir /etc/sysconfig/slurm \
        /var/spool/slurmd \
        /var/run/slurmd \
        /var/run/slurmdbd \
        /var/lib/slurmd \
        /var/log/slurm \
        /data \
    && touch /var/lib/slurmd/node_state \
        /var/lib/slurmd/front_end_state \
        /var/lib/slurmd/job_state \
        /var/lib/slurmd/resv_state \
        /var/lib/slurmd/trigger_state \
        /var/lib/slurmd/assoc_mgr_state \
        /var/lib/slurmd/assoc_usage \
        /var/lib/slurmd/qos_usage \
        /var/lib/slurmd/fed_mgr_state \
    && chown -R slurm:slurm /var/*/slurm* \
    && /sbin/create-munge-key

# Install envsubst, used by docker-entrypoint.sh
RUN yum install -y gettext

COPY slurm.conf /etc/slurm/slurm.conf.template
COPY slurmdbd.conf /etc/slurm/slurmdbd.conf.template
COPY burst_buffer.conf /etc/slurm/burst_buffer.conf

# Download and install etcd client
ARG ETCD_VERSION=3.3.13
ARG ETCD_DOWNLOAD_URL="https://github.com/etcd-io/etcd/releases/download/v$ETCD_VERSION/etcd-v$ETCD_VERSION-linux-amd64.tar.gz"

RUN set -x \
    && wget -O etcd.tar.gz "$ETCD_DOWNLOAD_URL" \
    && mkdir /usr/local/src/etcd \
    && tar xf etcd.tar.gz -C /usr/local/src/etcd --strip-components=1 \
    && install -D -m755 /usr/local/src/etcd/etcdctl /usr/local/bin/etcdctl \
    && rm etcd.tar.gz \
    && rm -rf /usr/local/src/etcd

# TODO: may want a separate dwstat binary
COPY bin/data-acc.tgz /usr/local/bin/
RUN set -x \
    && mkdir /usr/local/bin/data-acc \
    && tar xf /usr/local/bin/data-acc.tgz -C /usr/local/bin/data-acc \
    && install -D -m755 /usr/local/bin/data-acc/bin/dacd /usr/local/bin/dacd \
    && install -D -m755 /usr/local/bin/data-acc/bin/dacctl /usr/local/bin/dacctl \
    && install -D -m755 /usr/local/bin/dacctl /opt/cray/dw_wlm/default/bin/dw_wlm_cli \
    && install -D -m755 /usr/local/bin/dacctl /opt/cray/dws/default/bin/dwstat \
    && mkdir -p /var/lib/data-acc/ \
    && cp -r /usr/local/bin/data-acc/fs-ansible /var/lib/data-acc/ \
    && cd /var/lib/data-acc/fs-ansible \
    && virtualenv .venv
# TODO: need lots more work to get ansible running in here
#    && . .venv/bin/activate \
#    && pip install -U pip \
#    && pip install -U ansible \
#    && deactivate

RUN touch /var/log/dacctl.log \
    && chown slurm /var/log/dacctl.log \
    && chgrp slurm /var/log/dacctl.log

COPY docker-entrypoint.sh /usr/local/bin/docker-entrypoint.sh

ENTRYPOINT ["/usr/local/bin/docker-entrypoint.sh"]

CMD ["slurmdbd"]
