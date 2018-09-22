set -e
if [ $# -ne 1 ]; then
    exec php -a
else
    mv "$1" script.php
    exec php script.php
fi
