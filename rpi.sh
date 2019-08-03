#!/bin/bash

HOST="localhost"
USER="pi"

if [ $# -gt 0 ] ; then
    HOST="$1"
fi
if [ $# -gt 1 ] ; then 
    USER="$2"
fi

echo "Bulding..."
cd "cmd/teak" || exit -1

GOARCH=arm64 GOOS=linux go install || exit -1

echo "Generated at $GOPATH/bin/linux_arm64"
ls "$GOPATH/bin/linux_arm64"

# scp "$GOPATH/bin/linux_arm64/teak" "$USER@$HOST:/opt/bin"
rsync -avz -e ssh "$GOPATH/bin/linux_arm64/teak" "$USER@$HOST:/opt/bin"

# ssh "$USER@$HOST" 'nohup "/opt/bin/teak" serve --port 8000 > console.log 2>&1 &'

ssh "$USER@$HOST" 'killall -9 teak ; "/opt/bin/teak" serve --port 8000'

