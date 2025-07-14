#!/usr/bin/env bash
set -o errexit

APP=${1:?app name as first argument}
FULLNAME=${2:?app name with version as second argument}

echo "Building linux binary $VER"
mkdir -p build
mkdir -p $FULLNAME
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -trimpath -installsuffix cgo -ldflags="-w -s" -o $FULLNAME/$APP ..
#echo "Release $VER ($(date)): $1" > $FULLNAME/$VER.txt
tar czf build/$FULLNAME.tar.gz -C $FULLNAME .
rm -vr $FULLNAME
echo "Built grok-telegram-bot binary for linux at build/$FULLNAME.tar.gz"
