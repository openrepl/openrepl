FROM debian:jessie

RUN apt-get update && apt-get -y install gforth

ADD runforth.sh runforth.sh
ENTRYPOINT ["bash", "runforth.sh"]
