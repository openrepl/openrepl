set -e
if [ $# -ne 1 ]; then
    exec cling
else
    mv "$1" code.cpp
    clang++ code.cpp
    chmod 700 a.out
    exec ./a.out
fi
