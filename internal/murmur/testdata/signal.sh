#!/bin/sh

set -e

trap catch 2 # SIGINT

catch()
{
    echo "Caught SIGINT"
    sleep 2
    echo "Graceful shutdown OK"
}

echo "Started"
cat
