#!/bin/bash

EVM="./go-ethereum/build/bin/evm"
EVM_LOG="evm-t8n-substate.log"

ws="1 4 16 64 8 12 24 32 48 2 44 88"
ss=$(seq 1 9)
for w in $ws
do
    for s in $ss
    do
        b=$(expr ${s} - 1)
        log="evm-t8n-substate-w${w}-${b}-${s}M.log"
        echo "$log"
        # if log file exists
        if [[ -f "$log" ]]; then
            continue
        fi
        $EVM t8n-substate ${b}000001 ${s}000000 --workers ${w} 2>&1 | tee $EVM_LOG
        mv $EVM_LOG $log
    done
done
