FROM php:7.0-alpine

RUN apk --no-cache add bash
ADD runphp.sh runphp.sh
ENTRYPOINT ["bash","runphp.sh"]
