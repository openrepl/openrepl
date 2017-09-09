FROM ubuntu:xenial

RUN apt-get update && apt-get install -y curl bzip2                                                 \
    && curl https://root.cern.ch/download/cling//cling_2017-09-09_ubuntu16.tar.bz2 > cling.tar.bz2  \
    && tar -xf cling.tar.bz2                                                                        \
    && rm cling.tar.bz2

ADD runcpp.sh runcpp.sh
ENTRYPOINT ["bash","runcpp.sh"]
