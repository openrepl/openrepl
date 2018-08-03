if [ $# -ne 1 ]; then
    gore
else
    cp "$1" /code.go
    go run /code.go
fi
