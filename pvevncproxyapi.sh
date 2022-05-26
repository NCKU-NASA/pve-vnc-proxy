#!/bin/bash

if [ $# -lt 1 ]
then
    echo "usage: $0 <command>"
    exit 1
fi

workdir="/etc/pvevncproxy"

case $1 in
    start)
        . ./venv/bin/activate
        gunicorn --bind 127.0.0.1:4001 vnc:app
        ;;
    reload)
        > .reload
esac
