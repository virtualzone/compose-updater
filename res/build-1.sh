#!/bin/sh
BASEDIR=$(dirname $(realpath "$0"))
docker build --no-cache -t watcher-test-1 $BASEDIR