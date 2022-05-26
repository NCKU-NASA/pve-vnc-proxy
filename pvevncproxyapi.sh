#!/bin/bash

workdir="/etc/pvevncproxy"

. ./venv/bin/activate

gunicorn --bind 127.0.0.1:4001 vnc:app

