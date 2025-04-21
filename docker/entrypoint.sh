#!/usr/bin/env sh

if [ "$1" = 'iriscli' ]; then
    shift
    exec iriscli --home=$IRIS_DIR "$@"
fi

exec irisd --home=$IRIS_DIR "$@"
