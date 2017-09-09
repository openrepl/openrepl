#!/bin/bash

NAME=$(mktemp -d -p /home XXXXXX)

function cleanup() {
    docker kill $(basename $NAME)
    rm -r $NAME
}

trap cleanup EXIT

set -x

if [ $# -eq 2 ]; then
    wget http://60s/get?id="$2" -O $NAME/script || exit 1
    DOCKERFLAGS=(-v "$NAME:$NAME")
    SARGS=($NAME/script)
fi

bash filter.sh "$1" || exit 1

docker run ${DOCKERFLAGS[*]} --name $(basename $NAME) -it --rm -m 64m openrepl/$1 $SARGS
