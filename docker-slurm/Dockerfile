FROM centos:7

LABEL org.label-schema.vcs-url="https://github.com/giovtorres/slurm-docker-cluster" \
      org.label-schema.docker.cmd="docker-compose up -d" \
      org.label-schema.name="slurm-docker-cluster" \
      org.label-schema.description="Slurm Docker cluster on CentOS 7" \
      maintainer="Giovanni Torres"

ARG SLURM_VERSION=17.02.9
ARG SLURM_DOWNLOAD_MD5=6bd0b38e6bf08f3426a7dd1e663a2e3c
ARG SLURM_DOWNLOAD_URL=https://download.schedmd.com/slurm/slurm-"$SLURM_VERSION".tar.bz2

ARG GOSU_VERSION=1.10

RUN yum makecache fast \
    && yum -y install epel-release \
    && yum -y install \
           wget \
           bzip2 \
           perl \
           gcc \
           gcc-c++\
           vim-enhanced \
           git \
           json-c \
           json-c-devel \
           make \
           munge \
           munge-devel \
           python-devel \
           python-pip \
           python34 \
           python34-devel \
           python34-pip \
           mariadb-server \
           mariadb-devel \
           psmisc \
           bash-completion \
    && yum clean all \
    && rm -rf /var/cache/yum

RUN pip install Cython nose \
    && pip3 install Cython nose

RUN set -x \
    && wget -O /usr/local/bin/gosu "https://github.com/tianon/gosu/releases/download/$GOSU_VERSION/gosu-amd64" \
#    && wget -O /usr/local/bin/gosu.asc "https://github.com/tianon/gosu/releases/download/$GOSU_VERSION/gosu-amd64" \
#    && export GNUPGHOME="$(mktemp -d)" \
#    && gpg --keyserver ha.pool.sks-keyservers.net --recv-keys B42F6819007F00F88E364FD4036A9C25BF357DD4 \
#    && gpg --batch --verify /usr/local/bin/gosu.asc /usr/local/bin/gosu \
#    && rm -rf $GNUPGHOME /usr/local/bin/gosu.asc \
    && chmod +x /usr/local/bin/gosu \
    && gosu nobody true

RUN groupadd -r slurm --gid=995 && useradd -r -g slurm --uid=995 slurm

RUN set -x \
    && wget -O slurm.tar.bz2 "$SLURM_DOWNLOAD_URL" \
    && echo "$SLURM_DOWNLOAD_MD5" slurm.tar.bz2 | md5sum -c - \
    && mkdir /usr/local/src/slurm \
    && tar jxf slurm.tar.bz2 -C /usr/local/src/slurm --strip-components=1 \
    && rm slurm.tar.bz2 \
    && cd /usr/local/src/slurm \
    && ./configure --enable-debug --prefix=/usr --sysconfdir=/etc/slurm \
        --with-mysql_config=/usr/bin  --libdir=/usr/lib64 \
    && make install \
    && install -D -m644 etc/cgroup.conf.example /etc/slurm/cgroup.conf.example \
    && install -D -m644 etc/slurm.conf.example /etc/slurm/slurm.conf.example \
    && install -D -m644 etc/slurm.epilog.clean /etc/slurm/slurm.epilog.clean \
    && install -D -m644 etc/slurmdbd.conf.example /etc/slurm/slurmdbd.conf.example \
    && install -D -m644 contribs/slurm_completion_help/slurm_completion.sh /etc/profile.d/slurm_completion.sh \
#    && install -D -m755 src/plugins/burst_buffer/cray/dw_wlm_cli /opt/cray/dw_wlm/default/bin/dw_wlm_cli \
#    && install -D -m755 src/plugins/burst_buffer/cray/dwstat /opt/cray/dws/default/bin/dwstat \
    && cd \
    && rm -rf /usr/local/src/slurm \
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

COPY slurm.conf /etc/slurm/slurm.conf
COPY slurmdbd.conf /etc/slurm/slurmdbd.conf
COPY burst_buffer.conf /etc/slurm/burst_buffer.conf

ARG BURSTBUFFER_BRANCH=etcd

# Install burstbuffer python wrapper
RUN yum -y install python34 python-pip python-wheel
RUN pip install -U pip
RUN pip install virtualenv
RUN cd /usr/local/src \
    && git clone https://github.com/johngarbutt/burstbuffer \
    && cd burstbuffer \
    && git checkout $BURSTBUFFER_BRANCH \
    && git checkout $BURSTBUFFER_BRANCH \
    && virtualenv -p python3 .venv \
    && . .venv/bin/activate \
    && pip install -e . \
    && deactivate \
    && mkdir -p /opt/cray/dw_wlm/default/bin \
    && ln -s /usr/local/src/burstbuffer/.venv/bin/fakewarp /opt/cray/dw_wlm/default/bin/dw_wlm_cli \
    && ln -s /usr/local/src/burstbuffer/.venv/bin/buffernode /usr/local/bin/buffernode


ARG ETCD_VERSION=3.3.1
ARG ETCD_DOWNLOAD_URL="https://github.com/coreos/etcd/releases/download/v$ETCD_VERSION/etcd-v$ETCD_VERSION-linux-amd64.tar.gz"

# Download and install etcd client
RUN set -x \
    && wget -O etcd.tar.gz "$ETCD_DOWNLOAD_URL" \
    && mkdir /usr/local/src/etcd \
    && tar xf etcd.tar.gz -C /usr/local/src/etcd --strip-components=1 \
    && install -D -m755 /usr/local/src/etcd/etcdctl /usr/local/bin/etcdctl \
    && rm etcd.tar.gz \
    && rm -rf /usr/local/src/etcd

COPY docker-entrypoint.sh /usr/local/bin/docker-entrypoint.sh
ENTRYPOINT ["/usr/local/bin/docker-entrypoint.sh"]

CMD ["slurmdbd"]
