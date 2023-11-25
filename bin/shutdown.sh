#!/bin/sh

root=`dirname $0`/..
kill `cat "${root}/var/server.pid"`
