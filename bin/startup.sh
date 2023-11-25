#!/bin/sh

root=`dirname $0`/..
. "${root}/.venv/bin/activate"
"${root}/build-local/sensors/bin/sensors" -config "${root}/config.json" >>"${root}/var/log/sensors.log" 2>&1 &
disown
