#!/bin/bash

PATH_SERVER=../../cmd/highloadcup2018/
PATH_SCRIPTS=../../scripts/
PATH_OUT=../../out/

mkdir -p $PATH_OUT

go build -o $PATH_OUT/highloadcup2018 $PATH_SERVER

cp $PATH_SCRIPTS/*.sh $PATH_OUT
cp $PATH_SCRIPTS/*.py $PATH_OUT

docker build -f Dockerfile -t highloadcup2018 ../../out/ && docker tag highloadcup2018 stor.highloadcup.ru/accounts/big_cheetah
# docker push stor.highloadcup.ru/accounts/big_cheetah
