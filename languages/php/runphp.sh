if [ $# -ne 1 ]; then
    php -a
else
    mv "$1" script.php
    php script.php
fi
