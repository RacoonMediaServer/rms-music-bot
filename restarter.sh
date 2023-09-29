#!/bin/bash

fifo_file=$1

while [[ 1 ]]
do
    read  < ${fifo_file}
    docker-compose restart music-bot
    docker-compose restart navidrome
done