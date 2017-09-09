if [ $# -ne 1 ]; then
    cling
else
    mv "$1" code.cpp
    clang++ code.cpp
    chmod 700 a.out
    ./a.out
fi
