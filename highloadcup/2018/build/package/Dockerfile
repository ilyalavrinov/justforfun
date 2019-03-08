FROM ubuntu:18.04

RUN apt-get update
RUN apt-get install --assume-yes apt-utils
RUN apt-get install --assume-yes python redis-server

RUN sed -i.bak  s/"bind.*"//g /etc/redis/redis.conf

WORKDIR /root

EXPOSE 80

ADD highloadcup2018 ./bin/
ADD run.sh ./bin/
ADD load_initial.py ./bin/

CMD ./bin/run.sh
