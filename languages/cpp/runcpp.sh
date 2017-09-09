if [ $# -ne 1 ]; then
    cling
else
    clang++ "$1"
    chmod 700 a.out
    ./a.out
fi
