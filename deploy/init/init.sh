#!/bin/bash

HOST=${1:?first argument must be remote hostname}
NAME=${2:?app name as second argument}
WORKDIR=/opt/$NAME

scp -r init/artifacts  $HOST:/tmp/artifacts
ssh $HOST "sudo mkdir -p /opt/$NAME"
ssh $HOST "sudo mv /tmp/artifacts/.env /opt/${NAME}/"

ssh $HOST "sudo mv /tmp/artifacts/${NAME}.service /etc/systemd/system/ && rm -rf /tmp/artifacts"
ssh $HOST "sudo systemctl enable $NAME.service"
ssh $HOST "sudo systemctl restart $NAME.service"
