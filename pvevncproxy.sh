#!/bin/bash

workdir="/etc/pvevncproxy"

. ./venv/bin/activate

python vncsock.py
