if [ $# -ne 1 ]; then
    echo -n 'NodeJS '
    node --version
fi
exec node -- $@
