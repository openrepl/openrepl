#!/bin/bash

NAME=$(mktemp -d)

function cleanup() {
    docker kill $(basename $NAME) &> /dev/null
    rm -rf $NAME
}

trap cleanup EXIT

DOCKERFLAGS=()

if [ $# -eq 2 ]; then
    wget -q http://60s/get?id="$2" -O $NAME/script || exit 1
    DOCKERFLAGS=(-v "$NAME:$NAME")
    SARGS=($NAME/script)
fi

bash filter.sh "$1" || exit 1

if [ "$1" == typescript ]; then
    DOCKERFLAGS+=(-m 128m)
else
    DOCKERFLAGS+=(-m 64m)
fi

docker run ${DOCKERFLAGS[*]} --name $(basename $NAME) -it --rm openrepl/$1 $SARGS 2> /dev/null
