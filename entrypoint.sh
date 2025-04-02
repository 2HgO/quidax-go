#!/bin/sh

echo 'creating tigerbeetle file';
[ ! -f /db/data/quidax_01.tigerbeetle ] && /db/tigerbeetle format --cluster=0 --development --replica=0 --replica-count=1 /db/data/quidax_01.tigerbeetle;

echo 'starting transaction db';
/db/tigerbeetle start --addresses='0.0.0.0:3000' --development --cache-grid=512MiB /db/data/quidax_01.tigerbeetle &
sleep 3 && /app/quidax-go;
