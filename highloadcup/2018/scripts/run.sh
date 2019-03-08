#!/bin/bash

PATH=$PATH:/root/bin/

service redis-server restart
load_initial.py /tmp/data/data.zip http://localhost:80 &
highloadcup2018
