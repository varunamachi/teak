#!/bin/bash

if [ $# -lt 1 ]; then 
    echo "Command name not provided"
    exit -1
fi

cmd=${1}
scriptDir="$( cd "$(dirname "$0")" || exit ; pwd -P )"
cmdDir="${scriptDir}/cmd/${cmd}"
if [ ! -d "$cmdDir" ] ; then
    echo "Command directory $cmdDir does not exist"
fi
cd "$cmdDir" || exit -1
echo "Building...."
go install || exit -1
echo "Running...."
echo
shift
"$cmd" "$@"



