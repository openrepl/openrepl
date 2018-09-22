if [ $# -ne 1 ]; then
    exec gforth
else
    exec gforth "$1" -e bye
fi
