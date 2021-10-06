#!/bin/sh

docker build -t srt-advanced-server .
docker run -it -p "11000-11010:11000-11010/tcp" -p "11000-11010:11000-11010/udp" srt-advanced-server