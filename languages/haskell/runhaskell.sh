set -e
if [ $# -ne 1 ]; then
    exec ghci
else
    mv "$1" /code.hs
    exec runghc -- -- /code.hs
fi
