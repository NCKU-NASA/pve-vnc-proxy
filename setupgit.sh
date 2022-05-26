#!/bin/bash
cd /tmp

rm -r pve-vnc-proxy

git clone https://github.com/NCKU-NASA/pve-vnc-proxy

cd pve-vnc-proxy

for a in $(ls -a)
do
    if [ "$a" != "." ] && [ "$a" != ".." ] && [ "$a" != ".git" ] && [ "$a" != "Readme.md" ] && [ "$a" != "install.sh" ] && [ "$a" != "remove.sh" ] && [ "$a" != "setupgit.sh" ] && [ "$a" != "nodes.yaml" ] && [ "$a" != "uservmlist.yaml" ]
    then
        rm -rf $a
    fi
done

for a in $(ls -a /etc/pvevncproxy)
do
    if [ "$a" != "." ] && [ "$a" != ".." ] && [ "$a" != "nodes.yaml" ] && [ "$a" != "uservmlist.yaml" ] && [ "$(cat /etc/pvevncproxy/.gitignore | sed 's/\/.*//g' | sed '/^!.*/d' | grep -P "^$(echo "$a" | sed 's/\./\\\./g')$")" == "" ]
    then
        sudo cp -r /etc/pvevncproxy/$a $a
    fi
done

sudo cp /etc/systemd/system/pvevncproxy.service pvevncproxy.service
sudo cp /etc/systemd/system/pvevncproxyapi.service pvevncproxyapi.service
