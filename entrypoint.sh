#!/bin/sh

[ ! -f /data/$CLUSTERNAME.tigerbeetle ] && /tigerbeetle format --cluster=0 --replica=$CLUSTERNUMBER --replica-count=3 /data/$CLUSTERNAME.tigerbeetle;

/tigerbeetle "$@";
