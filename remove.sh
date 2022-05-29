#!/bin/bash

sudo systemctl stop pvevncproxy.service
sudo systemctl stop pvevncproxyapi.service
sudo systemctl disable pvevncproxy.service
sudo systemctl disable pvevncproxyapi.service

sudo rm /etc/systemd/system/pvevncproxy.service
sudo rm /etc/systemd/system/pvevncproxyapi.service

for filename in requirements.txt vnc.py vncsock.py pvevncproxy.sh pvevncproxyapi.sh app nodes.yaml uservmlist.yaml .gitignore venv
do
	sudo rm -r /etc/pvevncproxy/$filename
done

sudo mv /etc/pvevncproxy/server.key .
sudo mv /etc/pvevncproxy/server.crt .

if [ "`ls /etc/pvevncproxy`" = "" ]
then
	rm -r /etc/pvevncproxy
fi

echo ""
echo ""
echo "PVE VNC Proxy Service remove.sh complete."

for filename in server.key server.crt
do
	echo "Your ${filename} is at $(pwd)/${filename}."
done

echo "If you don't need then, please delete then."
