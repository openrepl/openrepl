set -e
if [ $# -ne 1 ]; then
    exec gore
else
    cp "$1" /code.go
    exec go run /code.go
fi
