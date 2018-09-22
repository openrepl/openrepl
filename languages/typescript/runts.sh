set -e
if [ $# -ne 1 ]; then
    echo -n 'NodeJS (TS-Node) '
    node --version
    exec ts-node
else
    mv "$1" script.ts
    exec ts-node script.ts
fi
