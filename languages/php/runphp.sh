if [ $# -ne 1 ]; then
    php
else
    mv "$1" script.php
    php script.php
fi
