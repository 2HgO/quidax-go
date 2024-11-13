#!/bin/sh

[ ! -f /data/$CLUSTERNAME.tigerbeetle ] && /tigerbeetle format --cluster=0 --replica=$CLUSTERNUMBER --replica-count=$CLUSTERCOUNT /data/$CLUSTERNAME.tigerbeetle;

/tigerbeetle "$@";
