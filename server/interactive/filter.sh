#!/bin/bash

GOOD="false"


for lang in lua python2 python3 forth cpp bash; do
    if [ "$1" == "$lang" ]; then
        GOOD="true"
    fi
done

if [ "$GOOD" == "false" ]; then
    echo Illegal language $1
    exit 1
fi
