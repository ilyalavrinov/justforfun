#! /bin/bash

docker run --rm -p 8080:80 -v "$PWD/../test/data/data:/tmp/data:ro" -t highloadcup2018 
