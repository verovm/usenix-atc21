#!/bin/bash

EVM="./evm"
EVM_LOG="evm-value-graph.log"

# 1_000_001--1_010_000, 1_010_001--1_020_000, ..., 1_990_001--2_000_000
ss=$(seq 1010 10 2000)
for s in $ss
do
    b=$(expr ${s} - 10)
    log="evm-value-graph-${s}k.log"
    echo "$log"
    # if log file exists
    if [[ -f "$log" ]]; then
        continue
    fi
    $EVM t8n-substate ${b}001 ${s}000 --workers 1 2>&1 > $EVM_LOG
    mv $EVM_LOG $log
done

