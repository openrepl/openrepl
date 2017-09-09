if [ $# -ne 1 ]; then
    gforth
else
    gforth "$1" -e bye
fi
