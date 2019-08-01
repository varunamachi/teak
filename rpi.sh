#!/bin/bash

HOST="localhost"
USER="pi"

if [ $# -gt 0 ] ; then
    HOST="$1"
fi
if [ $# -gt 1 ] ; then 
    USER="$2"
fi

cd "cmd/teak" || exit -1
GOOS=linux GOARCH=arm go install || exit -1

echo "Generated at $GOPATH/bin/linux_arm"
ls "$GOPATH/bin/linux_arm"

scp "$GOPATH/bin/linux_arm/teak" "$USER@$HOST:/opt/bin"

#ssh <topi> nohup "/opt/bin/teak" serve --port "8000" > "console.log" 2>&1 &

