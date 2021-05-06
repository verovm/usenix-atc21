#!/bin/bash

nproc=$1
EVM="./go-ethereum/build/bin/evm"

for n in {0..8}; do
   #from=$(( n*1000000 + 1))
   #to=$(( (n+1) * 1000000 ))
   from=$(( n*10000 + 1 + 3000000))
   to=$(( (n+1) * 10000 + 3000000))
   $EVM t8n-substate $from $to --workers $nproc --skip-transfer-txs --skip-create-txs --log-file ${n}M.csv
done

python3 visualize.py
