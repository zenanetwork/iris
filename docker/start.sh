#!/usr/bin/env sh

# start processes
irisd start > ./logs/irisd.log &
irisd rest-server > ./logs/irisd-rest-server.log &
sleep 100
bridge start --all > ./logs/bridge.log &

# tail logs
tail -f ./logs/irisd.log ./logs/irisd-rest-server.log ./logs/bridge.log
