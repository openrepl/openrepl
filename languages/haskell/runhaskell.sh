if [ $# -ne 1 ]; then
    ghci
else
    mv "$1" /code.hs
    runghc -- -- /code.hs
fi
