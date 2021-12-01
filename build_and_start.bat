#!/bin/sh

docker build -t srt-advanced-server .
docker run -it -p "11000-11010:11000-11010/tcp" -p "11000-11010:11000-11010/udp" -p "11100-11110:11100-11110/tcp" -p "11100-11110:11100-11110/udp" srt-advanced-server