if [ $# -ne 1 ]; then
    ts-node
else
    mv "$1" script.ts
    ts-node script.ts
fi
