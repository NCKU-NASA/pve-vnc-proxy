#!/bin/bash

mkdir /etc/pvevncproxy
cp pvevncproxy /usr/local/bin
cp pvevncproxy.service /etc/systemd/system
for a in app
do
    cp -r $a /etc/pvevncproxy
done
for a in config.yaml
do
    if ! [ -f /etc/pvevncproxy/$a ]
    then
        cp -r $a /etc/pvevncproxy
    fi
done
